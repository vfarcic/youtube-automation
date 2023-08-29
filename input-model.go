package main

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	errMsg error
)

type inputModel struct {
	inputs     []textinput.Model
	focusIndex int
	err        error
}

func getInputFromString(question, answer string) (string, error) {
	qa := map[string]string{question: answer}
	output, err := getMultipleInputsFromString(qa)
	for _, value := range output {
		return value, err
	}
	return "", err
}

func getMultipleInputsFromString(qa map[string]string) (map[string]string, error) {
	p := tea.NewProgram(initialInputModel(qa), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return map[string]string{}, err
	}
	output := map[string]string{}
	for _, answer := range m.(inputModel).inputs {
		output[answer.Placeholder] = answer.Value()
	}
	return output, nil
}

func getInputFromBool(value bool) bool {
	return !value
}

func initialInputModel(qa map[string]string) inputModel {
	inputs := []textinput.Model{}
	for question, answer := range qa {
		i := textinput.New()
		i.Placeholder = question
		i.SetValue(answer)
		i.CharLimit = 100
		i.Width = 100
		if len(qa) == 1 || strings.HasPrefix(question, "1.") {
			i.Focus()
		}
		inputs = append(inputs, i)
	}
	return inputModel{
		inputs: inputs,
		err:    nil,
	}
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyTab, tea.KeyShiftTab, tea.KeyDown, tea.KeyUp:
			if msg.Type == tea.KeyShiftTab || msg.Type == tea.KeyUp {
				m.focusIndex--
			} else {
				m.focusIndex++
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					continue
				}
				m.inputs[i].Blur()
			}
		}
	case errMsg:
		m.err = msg
		return m, nil
	}
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m inputModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m inputModel) View() string {
	keys := make([]string, 0, len(m.inputs))
	for k := range m.inputs {
		keys = append(keys, m.inputs[k].Placeholder)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		for i2 := 0; i2 < len(m.inputs); i2++ {
			if m.inputs[i2].Placeholder == k {
				b.WriteString(m.inputs[i2].View())
				if i < len(m.inputs)-1 {
					b.WriteRune('\n')
				}
			}
		}
	}
	return b.String()
}
