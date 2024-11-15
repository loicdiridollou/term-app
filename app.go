package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	theme "github.com/loicdiridollou/term-app/theme"
)

type page int

type model struct {
	ready          bool
	switched       bool
	page           page
	theme_val      theme.Theme
	viewportWidth  int
	viewportHeight int
}

const (
	menuPage page = iota
	splashPage
	aboutPage
	faqPage
	shopPage
	paymentPage
	cartPage
	shippingPage
	confirmPage
	finalPage
)

func (m model) SwitchPage(page page) model {
	m.page = page
	m.switched = true
	return m
}

func (m model) AboutView() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Render(
			"1. # Amazingly awesome products for developers brought to you by a group of talented, good looking, and humble heroes...",
		),
		"",
		lipgloss.NewStyle().Render("2. # @thdxr"),
		"",
		lipgloss.NewStyle().Render("3. # @adamdotdev"),
		"",
	)
}

type CursorTickMsg struct{}

func (m model) CursorInit() tea.Cmd {
	return tea.Every(time.Millisecond*700, func(t time.Time) tea.Msg {
		return CursorTickMsg{}
	})
}

func (m model) SplashInit() tea.Cmd {
	return tea.Batch(m.CursorInit(), nil)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "a":
			m = m.SwitchPage(aboutPage)
			return m, nil
		case "m":
			m = m.SwitchPage(menuPage)
			return m, nil
		default:
			// any other key switches the screen
			return m, nil
		}
	default:
		return m, nil
	}
}

func (m model) SplashView() string {
	return lipgloss.Place(
		m.viewportWidth,
		m.viewportHeight,
		lipgloss.Center,
		lipgloss.Center,
		m.LogoView(),
	)
}

func (m model) View() string {
	switch m.page {
	case splashPage:
		return m.SplashView()
	case menuPage:
		return m.MenuView()
	case aboutPage:
		return m.AboutView()
	default:
		return ""
	}
}

func (m model) LogoView() string {
	return "terminal" + " Heloo new"
}

func (m model) MenuView() string {
	return "terminal" + " new menu"
}

func MenuPage() model {
	return model{page: splashPage, viewportWidth: 100, viewportHeight: 100}
}

func (m model) Init() tea.Cmd {
	return m.SplashInit()
}

func main() {
	p := tea.NewProgram(MenuPage(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error starting program:", err)
		os.Exit(1)
	}
}