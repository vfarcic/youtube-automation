package main

import (
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type choice struct {
	cursor   int
	selected string
	choices  []string
}

func getChoice(titles []string) string {
	p := tea.NewProgram(choice{choices: titles})
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
	return m.(choice).selected
}

func (m choice) Init() tea.Cmd {
	return nil
}

func (m choice) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "enter":
			m.selected = m.choices[m.cursor]
			return m, tea.Quit
		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}
		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		}
	}
	return m, nil
}

func (m choice) View() string {
	s := strings.Builder{}
	s.WriteString("Which video title do you prefer?\n\n")

	for i := 0; i < len(m.choices); i++ {
		if m.cursor == i {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(m.choices[i])
		s.WriteString("\n")
	}
	s.WriteString("\n(press enter to continue)\n")

	return s.String()
}
