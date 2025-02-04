package windows

import (
	"fmt"
	"strings"

	"github.com/Amrit02102004/RediCLI/utils"
	"github.com/gdamore/tcell/v2"

	"github.com/rivo/tview"
)

type EnhancedCommandSuggestion struct {
	command     string
	description string
	category    string
}


var enhancedCommandSuggestions = []EnhancedCommandSuggestion{
	{"see analytics", "Open analytics dashboard in browser", "Advanced"},
	{"flushall", "Delete all existing keys from Redis (USE WITH CAUTION)", "Advanced"},
	{"key filter set", "Open key set form with TTL in milliseconds", "Advanced"},
	{"key filter update", "Open key update form with KEEPTTL option", "Advanced"},
	{"get", "Retrieve the value of a key", "Basic"},
	{"set", "Set the string value of a key", "Basic"},
	{"del", "Delete a key", "Basic"},
	{"keys", "Find all keys matching a pattern", "Basic"},
	{"ttl", "Get the time to live for a key", "Basic"},
	{"expire", "Set a key's time to live in seconds", "Basic"},
	{"import", "Import data from CSV/XLSX file", "Data Management"},
	{"export", "Export data to CSV file", "Data Management"},
	{"get help", "Display help information and available commands", "Help"},
	{"clear", "clear console screen", "Basic"},
}

func Win3(app *tview.Application, logDisplay *tview.TextView, redis *utils.RedisConnection) (*tview.Flex, *tview.TextView, *tview.InputField, *tview.Flex) {
	cmdFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Create suggestion display
	suggestionDisplay := tview.NewTextView().
		SetDynamicColors(true).
		SetTextColor(tcell.ColorGray).
		SetTextAlign(tview.AlignCenter)

	// Create key-value display
	kvDisplay := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	kvDisplay.SetBorder(true)
	DisplayWelcomeMessage(kvDisplay)
	// Create a container for forms
	formContainer := tview.NewFlex().SetDirection(tview.FlexRow)

	// Create command input field
	cmdInput := tview.NewInputField().
		SetLabel("> ").
		SetFieldWidth(0)

	currentSuggestionIndex := 0
	var currentSuggestions []EnhancedCommandSuggestion

	cmdInput.SetChangedFunc(func(text string) {
		currentSuggestions = enhancedFilterSuggestions(text)
		if len(currentSuggestions) > 0 {
			suggestionText := ""
			for _, sugg := range currentSuggestions {
				suggestionText += fmt.Sprintf("[gray]%s[white] - %s (%s)\n", sugg.command, sugg.description, sugg.category)
			}
			suggestionDisplay.SetText(suggestionText)
		} else {
			suggestionDisplay.SetText("")
		}
	})

	cmdInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab && len(currentSuggestions) > 0 {
			currentSuggestionIndex = (currentSuggestionIndex + 1) % len(currentSuggestions)
			cmdInput.SetText(currentSuggestions[currentSuggestionIndex].command)
			return nil
		}
		return event
	})

	// Capture the main flex container to be used for overlay returns
	mainFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	cmdInput.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}

		cmd := strings.TrimSpace(cmdInput.GetText())
		if cmd == "" {
			return
		}

		logDisplay.Write([]byte(fmt.Sprintf("> %s\n", cmd)))

		switch {
		case cmd == "see analytics":
			go func() {
				err := redis.ServeAnalytics()
				if err != nil {
					logDisplay.Write([]byte(fmt.Sprintf("[red]Analytics Server Error:[white] %v\n", err)))
				}
			}()
			logDisplay.Write([]byte("[green]Analytics server started on http://localhost:8080[white]\n"))
			return 

		case cmd == "flushall":
			// Add confirmation dialog
			modal := tview.NewModal().
				SetText("Are you sure you want to delete ALL keys?\nThis action cannot be undone!").
				AddButtons([]string{"Yes", "No"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					if buttonLabel == "Yes" {
						err := redis.FlushAll()
						if err != nil {
							logDisplay.Write([]byte(fmt.Sprintf("[red]Error:[white] %v\n", err)))
						} else {
							logDisplay.Write([]byte("[green]Successfully deleted all keys[white]\n"))
						}
						RefreshData(logDisplay, kvDisplay, redis)
					}
					app.SetRoot(mainFlex, true)
				})
			app.SetRoot(modal, false)
		case cmd == "clear":
			Clear(kvDisplay)
			return
		case cmd == "key filter set":
			form := KeyFilterSetForm(app, redis, logDisplay, kvDisplay, mainFlex, formContainer, cmdFlex, suggestionDisplay, cmdInput)
			formContainer.Clear()
			formContainer.AddItem(form, 0, 1, true)
			// formContainer.SetTitle(" Key Filter Set Form ")
			cmdFlex.RemoveItem(kvDisplay)
			cmdFlex.RemoveItem(suggestionDisplay)
			cmdFlex.RemoveItem(cmdInput)
			cmdFlex.RemoveItem(formContainer)

			cmdFlex.AddItem(formContainer, 0, 1, false)
			cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
			cmdFlex.AddItem(cmdInput, 1, 0, true)
			return
		case cmd == "key filter update":
			form := KeyFilterUpdateForm(app, redis, logDisplay, kvDisplay, mainFlex, formContainer, cmdFlex, suggestionDisplay, cmdInput)
			formContainer.Clear()
			cmdFlex.Clear()
			formContainer.AddItem(form, 0, 1, true)
			// formContainer.SetTitle(" Key Filter Update Form ")
			cmdFlex.RemoveItem(kvDisplay)
			cmdFlex.RemoveItem(suggestionDisplay)
			cmdFlex.RemoveItem(cmdInput)

			cmdFlex.AddItem(formContainer, 0, 1, false)
			cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
			cmdFlex.AddItem(cmdInput, 1, 0, true)

			return
		case cmd == "import":
			form := ImportForm(app, redis, kvDisplay, logDisplay, cmdFlex, formContainer, suggestionDisplay, cmdInput)
			formContainer.Clear()
			cmdFlex.Clear()
			formContainer.AddItem(form, 0, 1, true)
			cmdFlex.AddItem(formContainer, 0, 1, false)
			cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
			cmdFlex.AddItem(cmdInput, 1, 0, true)
			return

		case cmd == "export":
			form := ExportForm(app, redis, kvDisplay, logDisplay, cmdFlex, formContainer, suggestionDisplay, cmdInput)
			formContainer.Clear()
			cmdFlex.Clear()
			formContainer.AddItem(form, 0, 1, true)
			cmdFlex.AddItem(formContainer, 0, 1, false)
			cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
			cmdFlex.AddItem(cmdInput, 1, 0, true)
			return
		case cmd == "get help":
			DisplayHelp(kvDisplay)
			return
		default:
			result, err := redis.ExecuteCommand(cmd)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Error:[white] %v\n", err)))
			} else {
				logDisplay.Write([]byte(fmt.Sprintf("[green]Result:[white] %v\n", result)))
			}
		}

		cmdInput.SetText("")
		RefreshData(logDisplay, kvDisplay, redis)
	})

	// Add suggestion display and command input to the flex container
	cmdFlex.AddItem(kvDisplay, 0, 1, false)
	cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
	cmdFlex.AddItem(cmdInput, 1, 0, true)

	return cmdFlex, kvDisplay, cmdInput, mainFlex
}
