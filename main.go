package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"torry/torrentfile"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ff4500"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5555"))

	subtitleStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#ff4500"))

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#6272A4"))

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f8f8f2")).
			Padding(0, 1).
			Align(lipgloss.Center)
)

const rawHeader = `
████████╗ ██████╗ ██████╗ ██████╗ ██╗   ██╗
╚══██╔══╝██╔═══██╗██╔══██╗██╔══██╗╚██╗ ██╔╝
   ██║   ██║   ██║██████╔╝██████╔╝ ╚████╔╝
   ██║   ██║   ██║██╔══██╗██╔══██╗  ╚██╔╝
   ██║   ╚██████╔╝██║  ██║██║  ██║   ██║
   ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝
`

type progressMsg float64

type model struct {
	err          error
	tf           torrentfile.TorrentFile
	selectedTf   string
	started      bool
	progress     float64
	progressChan chan float64
	buffChan     chan []byte
	progressBar  progress.Model
	spinner      spinner.Model
}

type clearErrorMsg struct{}

func clearError() tea.Cmd {
	return func() tea.Msg {
		return clearErrorMsg{}
	}
}

func watchProgress(ch chan float64) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-ch
		if !ok {
			return nil
		}
		return progressMsg(p)
	}
}

func initialModel(filepath string) model {
	tf, err := torrentfile.OpenTorrentFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	pc := make(chan float64, 100)
	bc := make(chan []byte)
	s := spinner.New()
	s.Spinner = spinner.Dot

	return model{
		tf:           tf,
		selectedTf:   filepath,
		started:      false,
		progress:     0,
		progressChan: pc,
		buffChan:     bc,
		progressBar:  progress.New(progress.WithDefaultGradient()),
		spinner:      s,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("TORRY"), m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit

		case "c":
			return m, clearError()

		case "d":
			if !m.started {
				m.started = true

				go func() {
					for range m.buffChan {
					}
				}()

				go func() {
					err := m.tf.D2f(&m.progressChan, &m.buffChan)
					if err != nil {
						log.Fatal(err)
					}
					close(m.progressChan)
				}()

				return m, watchProgress(m.progressChan)
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progressMsg:
		m.progress = float64(msg)
		cmd := m.progressBar.SetPercent(float64(msg) / 100)
		return m, tea.Batch(watchProgress(m.progressChan), cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.progressBar.Update(msg)
		m.progressBar = progressModel.(progress.Model)
		return m, cmd

	case clearErrorMsg:
		m.err = nil
	}

	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(rawHeader))
	b.WriteString("\n\n")

	if m.err != nil {
		errText := fmt.Sprintf("Error: %s", m.err.Error())
		b.WriteString(errorStyle.Render(errText))
		b.WriteString("\n\n")
	}

	sub := "A bittorrent client"
	b.WriteString(subtitleStyle.Render(sub))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Selected Torrent File: "))
	b.WriteString(m.selectedTf + "\n\n")

	b.WriteString(labelStyle.Render("Announce URL: "))
	b.WriteString(m.tf.Announce + "\n\n")

	b.WriteString(labelStyle.Render("Name: "))
	b.WriteString(m.tf.Name + "\n\n")

	b.WriteString(labelStyle.Render("Length: "))
	b.WriteString(strconv.Itoa(m.tf.Length) + "\n\n")

	b.WriteString(labelStyle.Render("Pieces: "))
	b.WriteString(strconv.Itoa(len(m.tf.PieceHashes)) + "\n\n")

	if !m.started {
		hint := "Press 'd' to start download"
		b.WriteString(hintStyle.Render(hint))
		b.WriteString("\n\n")
	} else {
		progressLine := m.spinner.View() + " Downloading: " + m.progressBar.View()
		b.WriteString(subtitleStyle.Render(progressLine))
		b.WriteString("\n\n")
	}

	b.WriteString("\n\n\n")
	footerText := "␣d␣ Start   ␣q␣ / ␣esc␣ Quit"
	b.WriteString(footerStyle.Render(footerText))

	return b.String()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Use: go run main.go path/to/some.torrent")
		os.Exit(1)
	}
	inputPath := os.Args[1]

	p := tea.NewProgram(initialModel(inputPath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Goodbye!")
}
