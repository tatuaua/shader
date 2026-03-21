package main

import (
	"bubbletea/test/shader"
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	speed    int
	mod1     float64
	mod2     float64
	mod3     float64
	selected int
	quitting bool
	frame    int
	text     string
	style    lipgloss.Style
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		m.text = ""
		m.frame += m.speed + 1

		image := shader.WrapSlice(shader.RenderFrame(float64(m.frame)*0.05, m.mod1, m.mod2, m.mod3))

		for _, column := range image {
			m.text += "\n"
			for _, row := range column {
				s := m.style.Background(lipgloss.RGBColor{R: row.R, G: row.G, B: row.B})
				m.text += s.Render("")
			}
		}

		return m, doTick()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
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

	labels := []string{
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
	return tea.View{Content: m.text + "\n" + status}
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Urgh:", err)
		os.Exit(1)
	}
}
