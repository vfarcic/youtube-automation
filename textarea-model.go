package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func getInputFromTextArea(question, value string, height int) string {
	p := tea.NewProgram(initialTextAreaModel(question, value, height), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
	return m.(textAreaModel).input.Value()
}

type textAreaModel struct {
	input    textarea.Model
	question string
	err      error
}

func initialTextAreaModel(question, value string, height int) textAreaModel {
	ti := textarea.New()
	ti.ShowLineNumbers = false
	ti.SetWidth(100)
	ti.SetHeight(20)
	ti.MaxWidth = 100
	ti.MaxHeight = height
	ti.CharLimit = 0
	ti.SetValue(value)
	ti.Focus()
	return textAreaModel{
		input:    ti,
		question: question,
		err:      nil,
	}
}

func (m textAreaModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m textAreaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.input.Focused() {
				m.input.Blur()
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		default:
			if !m.input.Focused() {
				cmd = m.input.Focus()
				cmds = append(cmds, cmd)
			}
		}
	case errMsg:
		m.err = msg
		return m, nil
	}
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m textAreaModel) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		m.question,
		m.input.View(),
		"(ctrl+c to continue)",
	) + "\n\n"
}

func modifyTextArea(value, header, errorMessage string) (string, error) {
	if len(value) == 0 && len(errorMessage) > 0 {
		return value, fmt.Errorf(redStyle.Render(errorMessage))
	}
	return strings.TrimSpace(getInputFromTextArea(header, value, 20)), nil
}

func modifyDescriptionTagsX(tags, descriptionTags, header, errorMessage string) (string, error) {
	if len(tags) == 0 {
		return descriptionTags, fmt.Errorf(redStyle.Render(errorMessage))
	}
	if len(descriptionTags) == 0 {
		descriptionTags = fmt.Sprintf("#%s", tags)
		descriptionTags = strings.ReplaceAll(descriptionTags, " ", "")
		descriptionTags = strings.ReplaceAll(descriptionTags, ",", " #")
	}
	return getInputFromTextArea(header, descriptionTags, 20), nil
}
