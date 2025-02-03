// win2.go
package windows

import "github.com/rivo/tview"

func (w *WindowManager) CreateLogView() *tview.TextView {
	w.logView = tview.NewTextView()
	w.logView.SetScrollable(true)
	w.logView.SetDynamicColors(true)
	w.logView.SetWrap(true)
	w.logView.SetBorder(true)
	w.logView.SetTitle(" Logs ")
	w.logView.Box.SetTitleAlign(tview.AlignLeft)

	w.logView.SetChangedFunc(func() {
		w.app.QueueUpdateDraw(func() {
			w.logView.ScrollToEnd()
		})
	})

	return w.logView
}