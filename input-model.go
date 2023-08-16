package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	errMsg error
)

type inputModel struct {
	input    textinput.Model
	question string
	err      error
}

func getInput(question, value string) string {
	p := tea.NewProgram(initialInputModel(question, value))
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
	return m.(inputModel).input.Value()
}

func initialInputModel(question, value string) inputModel {
	ti := textinput.New()
	ti.SetValue(value)
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 100
	return inputModel{
		input:    ti,
		question: question,
		err:      nil,
	}
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case errMsg:
		m.err = msg
		return m, nil
	}
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	return fmt.Sprintf(
		"%s\n%s",
		m.question,
		m.input.View(),
	) + "\n"
}
