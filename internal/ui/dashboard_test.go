package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestDashboardHandleInput(t *testing.T) {
	dash := NewDashboard()
	dash.SetupInputCapture()

	t.Run("searchField focused: shortcut keys should not be intercepted", func(t *testing.T) {
		dash.app.SetFocus(dash.searchField)

		shortcutRunes := []rune{'l', 'L', 'a', 'A', 'd', 'D', 'f', '/', 'q'}
		for _, r := range shortcutRunes {
			event := tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
			res := dash.handleInput(event)
			if res != event {
				t.Errorf("expected key event %q to be passed through when searchField is focused, but got %v", r, res)
			}
		}
	})

	t.Run("logsView focused: shortcut keys should be intercepted", func(t *testing.T) {
		dash.app.SetFocus(dash.logsView)

		lEvent := tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone)
		resL := dash.handleInput(lEvent)
		if resL != nil {
			t.Errorf("expected 'l' event to be intercepted (nil), got %v", resL)
		}

		aEvent := tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone)
		resA := dash.handleInput(aEvent)
		if resA != nil {
			t.Errorf("expected 'a' event to be intercepted (nil), got %v", resA)
		}
		if !dash.awsViewMode {
			t.Errorf("expected awsViewMode to be true after 'a' key press")
		}

		slashEvent := tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone)
		resSlash := dash.handleInput(slashEvent)
		if resSlash != nil {
			t.Errorf("expected '/' event to be intercepted (nil), got %v", resSlash)
		}
		if dash.app.GetFocus() != dash.searchField {
			t.Errorf("expected searchField to be focused after '/' key press")
		}
	})

	t.Run("simulate typing laravel when searchField is focused", func(t *testing.T) {
		dash := NewDashboard()
		dash.SetupInputCapture()
		dash.app.SetFocus(dash.searchField)

		handler := dash.searchField.InputHandler()
		word := "laravel"
		for _, r := range word {
			event := tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
			ev := dash.handleInput(event)
			if ev != nil {
				handler(ev, func(p tview.Primitive) {
					dash.app.SetFocus(p)
				})
			}
		}
		if dash.searchField.GetText() != "laravel" {
			t.Errorf("expected searchField text to be 'laravel', got %q", dash.searchField.GetText())
		}
	})

	t.Run("simulate pressing / then typing laravel from logsView", func(t *testing.T) {
		dash := NewDashboard()
		dash.SetupInputCapture()
		dash.app.SetFocus(dash.logsView)

		slashEv := tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone)
		ev := dash.handleInput(slashEv)
		if ev != nil {
			dash.logsView.InputHandler()(ev, func(p tview.Primitive) { dash.app.SetFocus(p) })
		}

		if dash.app.GetFocus() != dash.searchField {
			t.Fatalf("expected searchField to be focused after '/', got %v", dash.app.GetFocus())
		}

		word := "laravel"
		for _, r := range word {
			event := tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
			ev := dash.handleInput(event)
			if ev != nil {
				dash.searchField.InputHandler()(ev, func(p tview.Primitive) {
					dash.app.SetFocus(p)
				})
			}
		}

		if dash.searchField.GetText() != "laravel" {
			t.Errorf("expected searchField text to be 'laravel', got %q", dash.searchField.GetText())
		}
	})
}
