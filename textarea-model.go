package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func getTextArea(value string) string {
	p := tea.NewProgram(initialTextAreaModel(value))
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
	return m.(textAreaModel).input.Value()
}

type textAreaModel struct {
	input textarea.Model
	err   error
}

func initialTextAreaModel(value string) textAreaModel {
	ti := textarea.New()
	ti.ShowLineNumbers = false
	ti.SetWidth(100)
	ti.SetHeight(20)
	ti.MaxWidth = 100
	ti.MaxHeight = 20
	ti.CharLimit = 0
	ti.SetValue(value)
	ti.Focus()
	return textAreaModel{
		input: ti,
		err:   nil,
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
		"Tell me a story.\n%s\n\n%s",
		m.input.View(),
		"(ctrl+c to continue)",
	) + "\n\n"
}
