package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
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

		// Test 'l' (cycle log level)
		lEvent := tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone)
		resL := dash.handleInput(lEvent)
		if resL != nil {
			t.Errorf("expected 'l' event to be intercepted (nil), got %v", resL)
		}

		// Test 'a' (switch to AWS view)
		aEvent := tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone)
		resA := dash.handleInput(aEvent)
		if resA != nil {
			t.Errorf("expected 'a' event to be intercepted (nil), got %v", resA)
		}
		if !dash.awsViewMode {
			t.Errorf("expected awsViewMode to be true after 'a' key press")
		}

		// Test '/' (focus search field)
		slashEvent := tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone)
		resSlash := dash.handleInput(slashEvent)
		if resSlash != nil {
			t.Errorf("expected '/' event to be intercepted (nil), got %v", resSlash)
		}
		if dash.app.GetFocus() != dash.searchField {
			t.Errorf("expected searchField to be focused after '/' key press")
		}
	})
}
