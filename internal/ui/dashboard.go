// Package ui comment :)
package ui

import (
	"fmt"
	"time"

	"github.com/b92c/gowatch/internal/docker"
	"github.com/b92c/gowatch/internal/filter"
	"github.com/b92c/gowatch/pkg/metrics"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Dashboard struct {
	daemonError   error
	searchField   *tview.InputField
	logsView      *tview.TextView
	resourcesView *tview.TextView
	helpBar       *tview.TextView
	grid          *tview.Grid
	app           *tview.Application
	servicesTable *tview.Table
	daemonErrView *tview.TextView
	pages         *tview.Pages
	filterState   filter.FilterState
	userScrolling bool
	firstRender   bool
	filterMode    bool
}

func NewDashboard() *Dashboard {
	app := tview.NewApplication()

	servicesTable := tview.NewTable().
		SetBorders(true).
		SetFixed(1, 0)
	servicesTable.SetBorder(true).SetTitle(" Docker Services ").SetTitleAlign(tview.AlignLeft)

	logsView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	logsView.SetBorder(true).SetTitle(" Logs ").SetTitleAlign(tview.AlignLeft)

	resourcesView := tview.NewTextView().
		SetDynamicColors(true)
	resourcesView.SetBorder(true).SetTitle(" System Resources ").SetTitleAlign(tview.AlignLeft)

	helpBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[/][yellow] Search[white] | [f][yellow] Filter[white] | [l][yellow] Log Level[white] | [Esc][yellow] Clear[white] | [↑↓][yellow] Scroll[white] | [q][yellow] Quit[white]")
	helpBar.SetBorder(false).SetBackgroundColor(tcell.ColorBlack)

	searchField := tview.NewInputField().
		SetLabel("Search: ").
		SetPlaceholder("name, id or image...").
		SetFieldWidth(0)
	searchField.SetBorder(true).SetTitle(" Filter ")

	daemonErrView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	daemonErrView.SetBorder(true).
		SetTitle(" ⚠ Docker Daemon Offline ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorRed)

	errorHelpBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[q][yellow] Quit[white] | [gray]Waiting for Docker daemon...[-]")
	errorHelpBar.SetBorder(false).SetBackgroundColor(tcell.ColorBlack)

	grid := tview.NewGrid().
		SetRows(0, 0, 3, 0, 1).
		SetColumns(0).
		AddItem(servicesTable, 0, 0, 1, 1, 0, 0, false).
		AddItem(resourcesView, 1, 0, 1, 1, 0, 0, false).
		AddItem(searchField, 2, 0, 1, 1, 0, 0, false).
		AddItem(logsView, 3, 0, 1, 1, 0, 0, true).
		AddItem(helpBar, 4, 0, 1, 1, 0, 0, false)

	errorGrid := tview.NewGrid().
		SetRows(0, 1).
		SetColumns(0).
		AddItem(daemonErrView, 0, 0, 1, 1, 0, 0, false).
		AddItem(errorHelpBar, 1, 0, 1, 1, 0, 0, false)

	pages := tview.NewPages().
		AddPage("dashboard", grid, true, true).
		AddPage("error", errorGrid, true, false)

	app.SetRoot(pages, true)
	app.EnableMouse(true)
	app.SetFocus(logsView)

	dash := &Dashboard{
		app:           app,
		servicesTable: servicesTable,
		logsView:      logsView,
		resourcesView: resourcesView,
		helpBar:       helpBar,
		grid:          grid,
		pages:         pages,
		daemonErrView: daemonErrView,
		searchField:   searchField,
		filterState:   filter.NewFilterState(),
		userScrolling: false,
		firstRender:   true,
	}

	logsView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		dash.userScrolling = true
		return event
	})

	logsView.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseScrollUp || action == tview.MouseScrollDown {
			dash.userScrolling = true
		}
		return action, event
	})

	return dash
}

func (d *Dashboard) Update(containers docker.Containers) {
	filtered := filter.FilterContainers(containers, d.filterState)
	d.updateServicesTable(filtered)
	d.updateResourcesView(filtered.Host, filtered.C)
	d.updateLogsView(filtered)
}

func (d *Dashboard) SetupInputCapture() {
	d.searchField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			d.filterState.Clear()
			d.searchField.SetText("")
			d.app.SetFocus(d.logsView)
			d.filterMode = false
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			d.filterState.SetSearch(d.searchField.GetText())
			d.app.SetFocus(d.logsView)
			d.filterMode = false
			return nil
		}
		return event
	})
}

func (d *Dashboard) updateServicesTable(containers docker.Containers) {
	d.servicesTable.Clear()

	// Headers
	headers := []string{"Service", "State", "Image", "CPU %", "Memory", "Net Rx/Tx", "Net Pkts", "Disk R/W", "Disk Ops", "PIDs", "OOM", "Logs"}
	for i, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false)
		if i == 2 { // Image column
			cell.SetMaxWidth(25)
		}
		d.servicesTable.SetCell(0, i, cell)
	}

	// Data rows
	for row, c := range containers.C {
		serviceName := c.Service
		if serviceName == "" {
			serviceName = c.ID[:12]
		}

		stateColor := tcell.ColorGreen
		if c.State != "running" {
			stateColor = tcell.ColorRed
		}

		cpuSpark := tview.Escape(RenderSparkline(c.CPUHistory, 10))
		memSpark := tview.Escape(RenderSparkline(c.MemHistory, 10))

		memMB := fmt.Sprintf("%.2f MB", float64(c.MemUsage)/1024/1024)
		cpuStr := fmt.Sprintf("%s %.2f%%", cpuSpark, c.CPUPercent)
		memStr := fmt.Sprintf("%s %s", memSpark, memMB)

		netBytes := fmt.Sprintf("%s/%s", formatBytes(c.NetRxBytes), formatBytes(c.NetTxBytes))
		netPackets := fmt.Sprintf("%d/%d", c.NetRxPackets, c.NetTxPackets)
		diskBytes := fmt.Sprintf("%s/%s", formatBytes(c.DiskReadBytes), formatBytes(c.DiskWriteBytes))
		diskOps := fmt.Sprintf("%d/%d", c.DiskReadOps, c.DiskWriteOps)
		pids := fmt.Sprintf("%d", c.PIDsCurrent)
		oomEvents := fmt.Sprintf("%d", c.OOMEvents)
		logCount := fmt.Sprintf("%d lines", len(c.Log))
		oomColor := tcell.ColorGreen
		if c.OOMEvents > 0 {
			oomColor = tcell.ColorRed
		}

		cells := []struct {
			text  string
			color tcell.Color
		}{
			{serviceName, tcell.ColorWhite},
			{c.State, stateColor},
			{c.Image, tcell.ColorLightBlue},
			{cpuStr, tcell.ColorWhite},
			{memStr, tcell.ColorWhite},
			{netBytes, tcell.ColorWhite},
			{netPackets, tcell.ColorWhite},
			{diskBytes, tcell.ColorWhite},
			{diskOps, tcell.ColorWhite},
			{pids, tcell.ColorWhite},
			{oomEvents, oomColor},
			{logCount, tcell.ColorGray},
		}

		for col, cell := range cells {
			tableCell := tview.NewTableCell(cell.text).
				SetTextColor(cell.color).
				SetAlign(tview.AlignLeft)
			if col == 2 { // Image column
				tableCell.SetMaxWidth(25)
			}
			d.servicesTable.SetCell(row+1, col, tableCell)
		}
	}
}

func formatBytes(value uint64) string {
	const unit = 1024.0
	bytes := float64(value)
	if bytes < unit {
		return fmt.Sprintf("%d B", value)
	}
	if bytes < unit*unit {
		return fmt.Sprintf("%.1f KB", bytes/unit)
	}
	if bytes < unit*unit*unit {
		return fmt.Sprintf("%.1f MB", bytes/(unit*unit))
	}
	return fmt.Sprintf("%.1f GB", bytes/(unit*unit*unit))
}

func (d *Dashboard) updateResourcesView(host metrics.HostInfo, containers []docker.Container) {
	d.resourcesView.Clear()

	var totalCPU float64
	var totalMem uint64
	for _, c := range containers {
		totalCPU += c.CPUPercent
		totalMem += c.MemUsage
	}

	cpuCoresUsed := totalCPU / 100.0

	fmt.Fprintf(d.resourcesView, "[yellow]CPU Cores (Total):[-] %d\n", host.CPUCount)
	fmt.Fprintf(d.resourcesView, "[yellow]CPU Usage (Active):[-] %.2f cores (%.1f%%)\n\n", cpuCoresUsed, totalCPU)
	fmt.Fprintf(d.resourcesView, "[yellow]Memory Total:[-] %.2f GB\n", float64(host.MemTotal)/1024/1024/1024)
	fmt.Fprintf(d.resourcesView, "[yellow]Memory Used (Containers):[-] %.2f MB\n", float64(totalMem)/1024/1024)
	fmt.Fprintf(d.resourcesView, "[yellow]Memory Free (Available):[-] %.2f MB\n\n", float64(host.MemFree)/1024/1024)
	fmt.Fprintf(d.resourcesView, "[gray]Updated: %s[-]", time.Now().Format("15:04:05"))
}

var serviceColors = []string{
	"yellow", "cyan", "magenta", "green", "blue", "red",
	"darkcyan", "darkmagenta", "olive", "teal",
}

func (d *Dashboard) getServiceColor(serviceName string, containers docker.Containers) string {
	for i, c := range containers.C {
		name := c.Service
		if name == "" {
			if len(c.ID) >= 12 {
				name = c.ID[:12]
			} else {
				name = c.ID
			}
		}
		if name == serviceName {
			return serviceColors[i%len(serviceColors)]
		}
	}
	return "white"
}

func formatLogBadge(level metrics.LogLevel) string {
	switch level {
	case metrics.LogLevelFatal:
		return "[white:red:b][FATAL][-]"
	case metrics.LogLevelError:
		return "[red::b][ERROR][-]"
	case metrics.LogLevelWarn:
		return "[yellow::b][WARN][-]"
	case metrics.LogLevelInfo:
		return "[green][INFO][-]"
	case metrics.LogLevelDebug:
		return "[darkgray][DEBUG][-]"
	case metrics.LogLevelTrace:
		return "[darkgray][TRACE][-]"
	default:
		return ""
	}
}

func (d *Dashboard) updateLogsView(containers docker.Containers) {
	row, col := d.logsView.GetScrollOffset()

	title := " Logs "
	if d.filterState.MinLogLevel > metrics.LogLevelUnknown {
		title = fmt.Sprintf(" Logs [%s+] ", d.filterState.MinLogLevel.String())
	}
	d.logsView.SetTitle(title)

	d.logsView.Clear()
	for _, fl := range containers.FlatLogs {
		color := d.getServiceColor(fl.Service, containers)
		badge := formatLogBadge(fl.Level)
		if badge != "" {
			badge = badge + " "
		}
		fmt.Fprintf(d.logsView, "[yellow]%s[-] %s[%s]%s[-]\n", fl.Service, badge, color, tview.Escape(fl.Line))
	}

	if d.firstRender {
		d.logsView.ScrollToEnd()
		d.firstRender = false
	} else if !d.userScrolling {
		d.logsView.ScrollToEnd()
	} else {
		d.logsView.ScrollTo(row, col)
	}
}

func (d *Dashboard) Run() error {
	d.SetupInputCapture()
	d.app.SetInputCapture(d.handleInput)
	return d.app.Run()
}

func (d *Dashboard) Stop() {
	d.app.Stop()
}

func (d *Dashboard) handleInput(event *tcell.EventKey) *tcell.EventKey {
	if d.filterMode {
		return event
	}

	if event.Key() == tcell.KeyRune {
		switch event.Rune() {
		case 'q':
			d.app.Stop()
			return nil
		case '/', 'f':
			d.app.SetFocus(d.searchField)
			d.filterMode = true
			return nil
		case 'l', 'L':
			d.filterState.CycleMinLogLevel()
			return nil
		}
	}
	if event.Key() == tcell.KeyEscape {
		d.filterState.Clear()
		d.searchField.SetText("")
		d.app.SetFocus(d.logsView)
		d.filterMode = false
		return nil
	}
	return event
}

func (d *Dashboard) ShowDaemonError(err error) {
	d.app.QueueUpdateDraw(func() {
		d.daemonError = err
		d.daemonErrView.Clear()
		fmt.Fprintf(d.daemonErrView,
			"\n\n\n"+
				"    [red]██████╗  █████╗ ███████╗███╗   ███╗ ██████╗ ███╗   ██╗[-]\n"+
				"    [red]██╔══██╗██╔══██╗██╔════╝████╗ ████║██╔═══██╗████╗  ██║[-]\n"+
				"    [red]██║  ██║███████║█████╗  ██╔████╔██║██║   ██║██╔██╗ ██║[-]\n"+
				"    [red]██║  ██║██╔══██║██╔══╝  ██║╚██╔╝██║██║   ██║██║╚██╗██║[-]\n"+
				"    [red]██████╔╝██║  ██║███████╗██║ ╚═╝ ██║╚██████╔╝██║ ╚████║[-]\n"+
				"    [red]╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝[-]\n\n"+
				"    [red::b]⚠  Docker Daemon is not running[-]\n\n"+
				"    [white]GoWatch requires the Docker daemon to monitor containers.[-]\n"+
				"    [white]Please start Docker and GoWatch will reconnect automatically.[-]\n\n"+
				"    [gray]How to start Docker:[-]\n"+
				"    [yellow]  • macOS:[white]  Open Docker Desktop application[-]\n"+
				"    [yellow]  • Linux:[white]  sudo systemctl start docker[-]\n\n"+
				"    [gray]Error: %s[-]\n\n"+
				"    [darkgray]Retrying connection every 2 seconds...[-]",
			err.Error(),
		)
		d.pages.SwitchToPage("error")
	})
}

func (d *Dashboard) ClearDaemonError() {
	d.app.QueueUpdateDraw(func() {
		if d.daemonError == nil {
			return
		}
		d.daemonError = nil
		d.pages.SwitchToPage("dashboard")
	})
}
