package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	tea "charm.land/bubbletea/v2"
)

type model struct {
	speed       int
	mod1        float64
	mod2        float64
	mod3        float64
	selected    int
	quitting    bool
	frameNum    int
	text        []byte
	frameBuffer [Height][Width]RGB

	// FPS tracking
	lastTick    time.Time
	frameTimes  []time.Duration
	totalFrames int
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
		mod1: 0.7,
		mod2: 8.0,
		mod3: 0.2,
	}
}

func TextFromFrame(buf []byte, frame *[Height][Width]RGB) []byte {
	for y := range Height {
		buf = append(buf, '\n')
		for x := range Width {
			buf = append(buf, "\033[48;2;"...)
			buf = strconv.AppendUint(buf, uint64(frame[y][x].R), 10)
			buf = append(buf, ';')
			buf = strconv.AppendUint(buf, uint64(frame[y][x].G), 10)
			buf = append(buf, ';')
			buf = strconv.AppendUint(buf, uint64(frame[y][x].B), 10)
			buf = append(buf, "m  \033[0m"...)
		}
	}
	return buf
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		now := time.Time(msg)
		if !m.lastTick.IsZero() {
			dt := now.Sub(m.lastTick)
			m.frameTimes = append(m.frameTimes, dt)
			m.totalFrames++
		}
		m.lastTick = now

		m.text = m.text[:0]
		m.frameNum += m.speed + 1

		RenderFrame(float64(m.frameNum)*0.05, m.mod1, m.mod2, m.mod3, &m.frameBuffer)

		m.text = TextFromFrame(m.text, &m.frameBuffer)

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
	v := tea.View{Content: string(m.text) + "\n" + status}
	v.AltScreen = true
	return v
}

func main() {
	//f, _ := os.Create("cpu.prof")
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()
	p := tea.NewProgram(initialModel())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Urgh:", err)
		os.Exit(1)
	}

	m := finalModel.(model)
	const skip = 60
	if len(m.frameTimes) > skip*2 {
		trimmed := m.frameTimes[skip : len(m.frameTimes)-skip]
		var total time.Duration
		var maxFrame time.Duration
		for _, dt := range trimmed {
			total += dt
			if dt > maxFrame {
				maxFrame = dt
			}
		}
		avg := total / time.Duration(len(trimmed))
		fps := float64(time.Second) / float64(avg)
		target := time.Second / 60
		lag := avg - target
		if lag < 0 {
			lag = 0
		}
		droppedFrames := 0
		for _, dt := range trimmed {
			if dt > target*2 {
				droppedFrames++
			}
		}
		fmt.Printf("\nFPS Stats (%d frames, skipped first/last %d):\n", len(trimmed), skip)
		fmt.Printf("  avg fps:    %.1f\n", fps)
		fmt.Printf("  avg frame:  %.2fms\n", float64(avg.Microseconds())/1000.0)
		fmt.Printf("  worst:      %.2fms\n", float64(maxFrame.Microseconds())/1000.0)
		fmt.Printf("  avg lag:    %.2fms\n", float64(lag.Microseconds())/1000.0)
		fmt.Printf("  dropped:    %d (>2x target)\n", droppedFrames)
	} else {
		fmt.Printf("\nNot enough frames to report stats (need >%d, got %d)\n", skip*2, len(m.frameTimes))
	}
}
