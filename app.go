package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
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
	workPage
	projectPage
	blogPage
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
	cursor     cursorState
	splash     bool
	splashTime int
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
	// total := int64(1)
	// count := int64(3)

	bold := m.theme.TextAccent().Bold(true).Render
	accent := m.theme.TextAccent().Render
	base := m.theme.Base().Render
	// cursor := m.theme.Base().Background(m.theme.Highlight()).Render(" ")

	// menu := bold("m") + base(" ☰")
	// back := base("← ") + bold("esc") + base(" back")
	// mark := bold("t") + cursor
	logo := bold("loïc diridollou")
	about := accent("a") + base(" about")
	work := accent("w") + base(" work")
	project := accent("p") + base(" project")
	blog := accent("b") + base(" blog")
	contact := accent("c") + base(" contact")
	// accent(fmt.Sprintf(" $%2v", total/100)) +
	// base(fmt.Sprintf(" [%d]", count))

	switch m.page {
	case aboutPage:
		about = accent("a about")
	case workPage:
		work = accent("w work")
	case blogPage:
		work = accent("b blog")
	case projectPage:
		project = accent("p project")
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
		about,
		work,
		project,
		blog,
		contact,
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

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) SplashInit() tea.Cmd {
	return tea.Batch(tick(), m.CursorInit(), nil)
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
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tickMsg:
		m.state.splashTime -= 1
		if m.state.splashTime <= 0 {
			return m.SwitchPage(aboutPage), nil
		}
		return m, tick()

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
	}

	switch m.page {
	case splashPage:
		m, cmd = m.SplashUpdate(msg)
	}
	return m, cmd
}

type DelayCompleteMsg struct{}

func (m model) LoadCmds() []tea.Cmd {
	cmds := []tea.Cmd{}

	// Make sure the loading state shows for at least a couple seconds
	cmds = append(cmds, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return DelayCompleteMsg{}
	}))

	return cmds
}

func (m model) SplashUpdate(msg tea.Msg) (model, tea.Cmd) {
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
	return m.theme.TextAccent().Bold(true).Render("Loïc Diridollou") + m.CursorView()
}

func (m model) MenuView() string {
	return "terminal" + " new menu"
}

func RootPage() model {
	renderer := lipgloss.DefaultRenderer()
	return model{
		page:            splashPage,
		viewportWidth:   100,
		viewportHeight:  100,
		widthContainer:  100,
		heightContainer: 100,
		theme:           themer.BasicTheme(renderer, nil),
		renderer:        renderer,
		state:           state{splashTime: 2},
	}
}

func (m model) Init() tea.Cmd {
	return m.SplashInit()
}

func main_term() {
	p := tea.NewProgram(RootPage(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error starting program:", err)
		os.Exit(1)
	}
}

const (
	host = "localhost"
	port = "23234"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		panic("Too few arguments passed")
	}
	if args[1] == "terminal" {
		main_term()
	} else if args[1] == "server" {
		main_server()
	}
}

func main_server() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(), // Bubble Tea apps usually require a PTY.
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	renderer := bubbletea.MakeRenderer(s)

	m := model{
		page:            splashPage,
		viewportWidth:   100,
		viewportHeight:  100,
		widthContainer:  100,
		heightContainer: 100,
		theme:           themer.BasicTheme(renderer, nil),
		renderer:        renderer,
		state:           state{splashTime: 2},
	}

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}
