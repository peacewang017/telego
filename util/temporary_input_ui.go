package util

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TemporaryInputViewModelShared struct {
	set    bool
	header string
	tail   string
}

// Bubble Tea model
type TemporaryInputViewModel struct {
	textInput *textinput.Model
	Shared    *TemporaryInputViewModelShared
}

// Init initializes the model (required for Bubble Tea)
func (m TemporaryInputViewModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles user input
func (m TemporaryInputViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.Shared.set = true
			return m, tea.Quit
		case "ctrl+c":
			m.Shared.set = false
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	*m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View renders the UI
func (m TemporaryInputViewModel) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		m.Shared.header,
		// color.GreenString("为了进行自定义yml的管控，需要配置项目路径:"),
		m.textInput.View(),
		m.Shared.tail,
	)
}

func NewTemporaryInputViewModel(head string, placeholder string, tail string) *TemporaryInputViewModel {
	ti := textinput.New()
	ti.Placeholder = placeholder //"Enter your project directory"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40

	return &TemporaryInputViewModel{
		textInput: &ti,
		Shared:    &TemporaryInputViewModelShared{set: false, header: head, tail: tail},
	}
}

// StartTemporaryInputUI starts the Bubble Tea interface to get user input
// return (ok,content)
func StartTemporaryInputUI(head string, placeholder string, tail string) (bool, string) {

	m := NewTemporaryInputViewModel(head, placeholder, tail)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}

	return m.Shared.set, m.textInput.Value()
}
