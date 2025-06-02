package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const HEADER = `
████████╗ ██████╗ ██████╗ ██████╗ ██╗   ██╗
╚══██╔══╝██╔═══██╗██╔══██╗██╔══██╗╚██╗ ██╔╝
   ██║   ██║   ██║██████╔╝██████╔╝ ╚████╔╝ 
   ██║   ██║   ██║██╔══██╗██╔══██╗  ╚██╔╝  
   ██║   ╚██████╔╝██║  ██║██║  ██║   ██║   
   ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝                               
`

type model struct {
	message   string
	err       error
	input     textinput.Model
	sub       chan struct{}
	responses int
	spinner   spinner.Model
}

type clearErrorMsg struct{}
type responseMsg struct{}

func listenForActivity(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		for {
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(90)+100))
			sub <- struct{}{}
		}
	}
}

func waitForActivity(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-sub)
	}
}

func clearError() tea.Cmd {
	return func() tea.Msg {
		return clearErrorMsg{}
	}
}

func initialModel() model {
	input := textinput.New()
	input.Prompt = "aj>>> "
	input.Placeholder = "Your custom messge..."
	input.CharLimit = 250
	input.Width = 50
	return model{
		message: "Message",
		input:   input,
		sub:     make(chan struct{}),
		spinner: spinner.New(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("TORRY"), listenForActivity(m.sub),
		waitForActivity(m.sub), m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.input.Focused() {
			switch msg.String() {
			case "ctrl+c", "esc", "q":
				return m, tea.Quit
			case "e":
				if m.err != nil {
					m.err = nil
				} else {
					m.err = errors.New("damn son")
				}
			case "m":
				if m.message != "Bubbletea" {
					m.message = "Bubbletea"
				} else {
					m.message = "New Message"
				}
			case "c":
				return m, clearError()
			case "i":
				m.input.Focus()
				m.input.Reset()
			}
		} else {
			if msg.String() == "esc" {
				m.input.Reset()
				m.input.Blur()
			}
			if msg.String() == "enter" {
				m.message = m.input.Value()
				m.input.Reset()
				m.input.Blur()
			}
		}

	case clearErrorMsg:
		m.err = nil
	case responseMsg:
		m.responses++
		m.message = fmt.Sprintf("\n\nDownload %d%s", m.responses, "%")
		return m, waitForActivity(m.sub)
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	m.input, cmd = m.input.Update(msg)

	return m, cmd
}

func (m model) View() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(HEADER)
	s.WriteString("\n")
	if m.err != nil {
		s.WriteString("Error:")
		s.WriteString(m.err.Error())
	}
	s.WriteString("\nMessage: ")
	s.WriteString(m.message)
	s.WriteString("\n ")
	s.WriteString(m.spinner.View())
	s.WriteString(" Responses: ")
	s.WriteString(strconv.Itoa(m.responses))
	s.WriteString("\n\n\n")
	s.WriteString(m.input.View())
	return s.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Goodbye!")
}
