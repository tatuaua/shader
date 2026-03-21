package main

import (
	"bubbletea/test/shader"
	"fmt"
	"os"
	"strconv"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	speed    int
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
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		m.text = ""
		m.frame += m.speed + 1

		image := shader.WrapSlice(shader.RenderFrame(float64(m.frame) * 0.05))

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
		case "up":
			m.speed++
		case "down":
			m.speed--
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	var view tea.View
	if m.quitting {
		return view
	}

	return tea.View{Content: m.text + "\n" + strconv.Itoa(m.speed)}
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Urgh:", err)
		os.Exit(1)
	}
}
