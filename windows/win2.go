	// Create right side text view for logs

package windows

import (
		"github.com/rivo/tview"
)

func Win2(app *tview.Application) *tview.TextView {
    logs := tview.NewTextView().
        SetDynamicColors(true).
        SetChangedFunc(func() {
            app.Draw()
        })
    return logs
}