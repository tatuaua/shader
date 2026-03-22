package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	speed       int
	mod1        float64
	mod2        float64
	mod3        float64
	selected    int
	quitting    bool
	frameNum    int
	text        string
	style       lipgloss.Style
	frameBuffer [Height][Width]RGB
	shader      int // 0 = RenderFrame, 1 = RenderFrame2
}

type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return doTick()
}

func initialModel() model {
	return model{
		style: lipgloss.NewStyle().
			Width(2).
			Height(1).
			Background(lipgloss.Black),
		mod1: 0.7,
		mod2: 8.0,
		mod3: 0.2,
	}
}

func ColorText(text *strings.Builder, r, g, b int) {
	text.WriteString("\033[48;2;")
	text.Write(strconv.AppendUint(nil, uint64(r), 10))
	text.WriteByte(';')
	text.Write(strconv.AppendUint(nil, uint64(g), 10))
	text.WriteByte(';')
	text.Write(strconv.AppendUint(nil, uint64(b), 10))
	text.WriteString("m  \033[0m")
}

func (m model) TextFromFrame(frame *[Height][Width]RGB) string {
	var text strings.Builder
	text.Grow(Height * Width * 24)
	for y := range Height {
		text.WriteByte('\n')
		for x := range Width {
			ColorText(&text, int(frame[y][x].R), int(frame[y][x].G), int(frame[y][x].B))
		}
	}
	return text.String()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		m.text = ""
		m.frameNum += m.speed + 1

		switch m.shader {
		case 0:
			RenderFrame(float64(m.frameNum)*0.05, m.mod1, m.mod2, m.mod3, &m.frameBuffer)
		case 1:
			RenderFrame2(float64(m.frameNum)*0.05, &m.frameBuffer)
		}

		m.text = m.TextFromFrame(&m.frameBuffer)

		return m, doTick()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "tab":
			m.shader = (m.shader + 1) % 2
		case "right":
			if m.selected < 3 {
				m.selected++
			}
		case "left":
			if m.selected > 0 {
				m.selected--
			}
		case "up":
			switch m.selected {
			case 0:
				m.speed++
			case 1:
				m.mod1 += 0.1
			case 2:
				m.mod2 += 0.5
			case 3:
				m.mod3 += 0.05
			}
		case "down":
			switch m.selected {
			case 0:
				m.speed--
			case 1:
				m.mod1 -= 0.1
			case 2:
				m.mod2 -= 0.5
			case 3:
				m.mod3 -= 0.05
			}
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	var view tea.View
	if m.quitting {
		return view
	}

	shaderName := []string{"shader1", "shader2"}[m.shader]
	labels := []string{
		fmt.Sprintf("shader: %s", shaderName),
		fmt.Sprintf("speed: %d", m.speed),
		fmt.Sprintf("mod1: %.2f", m.mod1),
		fmt.Sprintf("mod2: %.2f", m.mod2),
		fmt.Sprintf("mod3: %.2f", m.mod3),
	}
	status := ""
	for i, l := range labels {
		if i == m.selected {
			status += "> " + l
		} else {
			status += "  " + l
		}
		if i < len(labels)-1 {
			status += "  "
		}
	}
	v := tea.View{Content: m.text + "\n" + status}
	v.AltScreen = true
	return v
}

func main() {
	f, _ := os.Create("cpu.prof")
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Urgh:", err)
		os.Exit(1)
	}
}
