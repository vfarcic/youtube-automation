package main

import (
	"fmt"
	"log"
	"strings"

	// TODO: Replace with Huh
	tea "github.com/charmbracelet/bubbletea"
)

var (
	// TODO: Remove
	errorMessage string
	// TODO: Remove
	confirmationMessage string
)

type choice struct {
	cursor   int
	question string
	selected map[int]string
	// selectedIndex int
	choices map[int]string
}

func getChoice(tasks map[int]Task, question string) (int, string) {
	choices := make(map[int]string)
	for key, item := range tasks {
		choices[key] = item.Title
	}
	p := tea.NewProgram(choice{choices: choices, question: question}, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
	for key, item := range m.(choice).selected {
		return key, item
	}
	return -1, ""
}

func getChoices(choices map[int]string, question string) map[int]string {
	p := tea.NewProgram(choice{choices: choices, question: question}, tea.WithAltScreen())
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
		case " ":
			if m.selected == nil {
				m.selected = make(map[int]string)
			}
			m.selected[m.cursor] = m.choices[m.cursor]
			errorMessage = ""
			confirmationMessage = ""
		case "enter":
			if m.selected == nil {
				m.selected = make(map[int]string)
			}
			m.selected[m.cursor] = m.choices[m.cursor]
			errorMessage = ""
			confirmationMessage = ""
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
	if len(errorMessage) > 0 {
		s.WriteString(errorStyle.Render(errorMessage))
	}
	if len(confirmationMessage) > 0 {
		s.WriteString(confirmationStyle.Render(confirmationMessage))
	}
	s.WriteString(fmt.Sprintf("%s\n\n", m.question))
	for i := 0; i < len(m.choices); i++ {
		if m.cursor == i {
			s.WriteString("(•) ")
		} else {
			selected := false
			for key := range m.selected {
				if key == i {
					s.WriteString("(•) ")
					selected = true
				}
			}
			if !selected {
				s.WriteString("( ) ")
			}
		}
		s.WriteString(m.choices[i])
		s.WriteString("\n")
	}
	s.WriteString("\n(press enter to continue)\n")
	return s.String()
}
