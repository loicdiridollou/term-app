package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	themer "github.com/loicdiridollou/term-app/theme"
)

type page int

type model struct {
	ready           bool
	switched        bool
	page            page
	theme           themer.Theme
	renderer        *lipgloss.Renderer
	viewportWidth   int
	viewportHeight  int
	widthContainer  int
	state           state
	heightContainer int
	size            size
	widthContent    int
	heightContent   int
	viewport        viewport.Model
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

type size = int

const (
	undersized size = iota
	small
	medium
	large
)

type cursorState struct {
	visible bool
}
type state struct {
	cursor cursorState
}

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

func (m model) HeaderView() string {
	total := int64(1)
	count := int64(3)

	bold := m.theme.TextAccent().Bold(true).Render
	accent := m.theme.TextAccent().Render
	base := m.theme.Base().Render
	// cursor := m.theme.Base().Background(m.theme.Highlight()).Render(" ")

	// menu := bold("m") + base(" ☰")
	// back := base("← ") + bold("esc") + base(" back")
	// mark := bold("t") + cursor
	logo := bold("terminal")
	shop := accent("s") + base(" shop")
	about := accent("a") + base(" about")
	faq := accent("f") + base(" faq")
	cart := accent("c") +
		base(" cart") +
		accent(fmt.Sprintf(" $%2v", total/100)) +
		base(fmt.Sprintf(" [%d]", count))

	switch m.page {
	case shopPage:
		shop = accent("s shop")
	case aboutPage:
		about = accent("a about")
	case faqPage:
		faq = accent("f faq")
	}

	// switch m.size {
	// case small:
	// 	tabs = []string{
	// 		mark,
	// 		cart,
	// 	}
	// case medium:
	// 	if m.hasMenu {
	// 		tabs = []string{
	// 			menu,
	// 			logo,
	// 			cart,
	// 		}
	// 	} else if m.checkout {
	// 		tabs = []string{
	// 			back,
	// 			logo,
	// 			cart,
	// 		}
	// 	} else {
	// 		tabs = []string{
	// 			logo,
	// 			cart,
	// 		}
	// 	}
	// default:
	// 	if m.checkout {
	// 		tabs = []string{
	// 			back,
	// 			logo,
	// 			cart,
	// 		}
	// 	} else {
	// 		tabs = []string{
	// 			logo,
	// 			shop,
	// 			about,
	// 			faq,
	// 			cart,
	// 		}
	// 	}
	// }

	tabs := []string{
		logo,
		shop,
		about,
		faq,
		cart,
	}

	return table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(m.renderer.NewStyle().Foreground(m.theme.Border())).
		Row(tabs...).
		Width(m.widthContainer).
		StyleFunc(func(row, col int) lipgloss.Style {
			return m.theme.Base().
				Padding(0, 1).
				AlignHorizontal(lipgloss.Center)
		}).
		Render()
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

var modifiedKeyMap = viewport.KeyMap{
	PageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdn", "page down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "page up"),
	),
	HalfPageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "½ page up"),
	),
	HalfPageDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "½ page down"),
	),
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
}

func (m model) updateViewport() model {
	headerHeight := lipgloss.Height(m.HeaderView())
	footerHeight := lipgloss.Height(m.FooterView())
	verticalMarginHeight := headerHeight + footerHeight + 2

	width := m.widthContainer - 4
	m.heightContent = m.heightContainer - verticalMarginHeight

	if !m.ready {
		m.viewport = viewport.New(width, m.heightContent)
		m.viewport.YPosition = headerHeight
		m.viewport.HighPerformanceRendering = false
		m.ready = true
	} else {
		m.viewport.Width = width
		m.viewport.Height = m.heightContent
		m.viewport.GotoTop()
	}

	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewportWidth = msg.Width
		m.viewportHeight = msg.Height

		switch {
		case m.viewportWidth < 20 || m.viewportHeight < 10:
			m.size = undersized
			m.widthContainer = m.viewportWidth
			m.heightContainer = m.viewportHeight
		case m.viewportWidth < 40:
			m.size = small
			m.widthContainer = m.viewportWidth
			m.heightContainer = m.viewportHeight
		case m.viewportWidth < 60:
			m.size = medium
			m.widthContainer = 40
			m.heightContainer = int(math.Min(float64(msg.Height), 30))
		default:
			m.size = large
			m.widthContainer = 100
			m.heightContainer = int(math.Min(float64(msg.Height), 60))
		}

		m.widthContent = m.widthContainer - 4
		m = m.updateViewport()
	case CursorTickMsg:
		m, cmd := m.CursorUpdate(msg)
		// TODO: this is bad, but otherwise the cursor doesn't blink
		return m, cmd
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
		case "h":
			m = m.SwitchPage(paymentPage)
			return m, nil
		default:
			// any other key switches the screen
			return m, nil
		}
	default:
		return m, nil
	}
	return m, nil
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

func (m model) FooterView() string {
	table := m.theme.Base().
		Width(m.widthContainer).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(m.theme.Border()).
		PaddingBottom(1).
		Align(lipgloss.Center)

	commands := []string{"test", "var", "none"}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		"free shipping on US orders over $40",
		table.Render(
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				commands...,
			),
		))
}

func (m model) getContent() string {
	page := "unknown"
	switch m.page {
	case aboutPage:
		page = m.AboutView()
	case splashPage:
		page = m.SplashView()
	case menuPage:
		page = m.MenuView()
	case paymentPage:
		page = m.HeaderView()
	}
	return page
}

func (m model) View() string {
	switch m.page {
	case splashPage:
		return m.SplashView()
	case menuPage:
		return m.MenuView()
	default:
		header := m.HeaderView()
		footer := m.FooterView()
		height := m.heightContainer
		height -= lipgloss.Height(header)
		height -= lipgloss.Height(footer)

		view := m.getContent()
		// return view
		child := lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			m.theme.Base().
				Width(m.widthContainer).
				Height(height).
				Padding(0, 1).
				Render(view),
			footer,
		)

		return m.renderer.Place(
			m.viewportWidth,
			m.viewportHeight,
			lipgloss.Center,
			lipgloss.Center,
			m.theme.Base().
				MaxWidth(m.widthContainer).
				MaxHeight(m.heightContainer).
				Render(child),
		)
	}
}

func (m model) CursorUpdate(msg tea.Msg) (model, tea.Cmd) {
	switch msg.(type) {
	case CursorTickMsg:
		m.state.cursor.visible = !m.state.cursor.visible
		return m, tea.Every(time.Millisecond*750, func(t time.Time) tea.Msg {
			return CursorTickMsg{}
		})
	}
	return m, nil
}

func (m model) CursorView() string {
	if m.state.cursor.visible {
		return m.theme.Base().Background(m.theme.Highlight()).Render(" ")
	} else {
		return m.theme.Base().Render(" ")
	}
}

func (m model) LogoView() string {
	return m.theme.TextAccent().Bold(true).Render("terminal") + m.CursorView()
}

func (m model) MenuView() string {
	return "terminal" + " new menu"
}

func MenuPage() model {
	renderer := lipgloss.DefaultRenderer()
	return model{
		page:            splashPage,
		viewportWidth:   100,
		viewportHeight:  100,
		widthContainer:  100,
		heightContainer: 100,
		theme:           themer.BasicTheme(renderer, nil),
		renderer:        renderer,
	}
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
