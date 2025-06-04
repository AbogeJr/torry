package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"torry/torrentfile"

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
	err        error
	tf         torrentfile.TorrentFile
	selectedTf string
	// progressChan chan float64
	// bufferChan   chan []byte
}

type clearErrorMsg struct{}

func clearError() tea.Cmd {
	return func() tea.Msg {
		return clearErrorMsg{}
	}
}

func initialModel(filepath string) model {
	tf, err := torrentfile.OpenTorrentFile(filepath)
	if err != nil {
		log.Fatal(err, tf)
	}

	return model{
		tf:         tf,
		selectedTf: filepath,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("TORRY"))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		case "e":
			if m.err != nil {
				m.err = nil
			} else {
				m.err = errors.New("damn son")
			}

		case "c":
			return m, clearError()

		}

	case clearErrorMsg:
		m.err = nil
	}
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
	s.WriteString("A torrent client ")
	s.WriteString("\n\n")
	s.WriteString("Selected Torrent File: ")
	s.WriteString(m.selectedTf)
	s.WriteString("\n")
	s.WriteString("Announce URL: ")
	s.WriteString(m.tf.Announce)
	s.WriteString("\n")
	s.WriteString("Name: ")
	s.WriteString(m.tf.Name)
	s.WriteString("\n")
	s.WriteString("Length: ")
	s.WriteString(strconv.Itoa(m.tf.Length))
	s.WriteString("\n")
	s.WriteString("Pieces: ")
	s.WriteString(strconv.Itoa(int(len(m.tf.PieceHashes))))
	s.WriteString("\n")

	s.WriteString("\n\n\n")
	// s.WriteString(m.input.View())
	return s.String()
}

func main() {
	inputPath := os.Args[1]
	p := tea.NewProgram(initialModel(inputPath))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Goodbye!")
}
