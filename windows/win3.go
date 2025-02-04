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
	{"clear all", "clear console and logs screen", "Basic"},
	{"clear logs", "clear logs screen", "Basic"},
	{"clear display", "clear display screen", "Basic"},
	{"add connection", "Open form to add and connect to a new Redis connection", "Connection Management"},
	{"get connections", "List all saved Redis connections", "Connection Management"},
	{"connect", "Connect to a saved Redis connection by name", "Connection Management"},
	{"del connection", "Delete a specific saved Redis connection", "Connection Management"},
	{"del all connections", "Delete all saved Redis connections", "Connection Management"},
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
	commandHistory := []string{}
	currentHistoryIndex := -1

	// Modify the SetInputCapture function in Win3
	cmdInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			if len(currentSuggestions) > 0 {
				currentSuggestionIndex = (currentSuggestionIndex + 1) % len(currentSuggestions)
				cmdInput.SetText(currentSuggestions[currentSuggestionIndex].command)
				return nil
			}
		case tcell.KeyUp:
			// Ensure we can access the last command
			if len(commandHistory) > 0 {
				if currentHistoryIndex == -1 {
					// First time pressing up, show the most recent command
					currentHistoryIndex = 0
				} else if currentHistoryIndex < len(commandHistory)-1 {
					// Move to the next older command
					currentHistoryIndex++
				}
				cmdInput.SetText(commandHistory[currentHistoryIndex])
			}
			return nil
		case tcell.KeyDown:
			// Navigate command history down
			if currentHistoryIndex > -1 {
				currentHistoryIndex--
				if currentHistoryIndex >= 0 {
					cmdInput.SetText(commandHistory[currentHistoryIndex])
				} else {
					currentHistoryIndex = -1
					cmdInput.SetText("")
				}
			}
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

		// Add command to history if it's not a repeat of the last command
		if len(commandHistory) == 0 || cmd != commandHistory[0] {
			commandHistory = append([]string{cmd}, commandHistory...)

			// Limit history size (optional)
			if len(commandHistory) > 50 {
				commandHistory = commandHistory[:50]
			}
		}

		// Reset history index
		currentHistoryIndex = -1

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
		case cmd == "clear all":
			Clear(kvDisplay, logDisplay, 1, 1)
			cmdFlex.Clear()
			cmdFlex.AddItem(kvDisplay, 0, 1, false)
			cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
			cmdFlex.AddItem(cmdInput, 1, 0, true)
			return
		case cmd == "clear logs":
			Clear(kvDisplay, logDisplay, 0, 1)
			return
		case cmd == "clear display":
			Clear(kvDisplay, logDisplay, 1, 0)
			cmdFlex.Clear()
			cmdFlex.AddItem(kvDisplay, 0, 1, false)
			cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
			cmdFlex.AddItem(cmdInput, 1, 0, true)
			return
		case cmd == "key filter set":
			form := KeyFilterSetForm(app, redis, logDisplay, kvDisplay, mainFlex, formContainer, cmdFlex, suggestionDisplay, cmdInput)
			formContainer.Clear()
			formContainer.AddItem(form, 0, 1, true)
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

		case strings.HasPrefix(cmd, "add connection"):
			form := ConnectionForm(app, logDisplay, redis, kvDisplay)
			formContainer.Clear()
			cmdFlex.Clear()
			formContainer.AddItem(form, 0, 1, true)
			cmdFlex.AddItem(formContainer, 0, 1, false)
			cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
			cmdFlex.AddItem(cmdInput, 1, 0, true)
			return

		case cmd == "get connections":
			connections, err := GetConnections()
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Error fetching connections: %v[white]\n", err)))
			} else {
				kvDisplay.SetText(FormatConnectionsList(connections))
			}
			return

		case strings.HasPrefix(cmd, "connect "):
			connectionName := strings.TrimSpace(strings.TrimPrefix(cmd, "connect"))
			config, err := FindConnectionByName(connectionName)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]%v[white]\n", err)))
				return
			}

			err = redis.Connect(config.Host, config.Port)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Connection failed: %v[white]\n", err)))
			} else {
				logDisplay.Write([]byte(fmt.Sprintf("[green]Connected to '%s' at %s:%s[white]\n",
					config.Name, config.Host, config.Port)))
				RefreshData(logDisplay, kvDisplay, redis)
			}
			return
		case strings.HasPrefix(cmd, "del connection "):
			connectionName := strings.TrimSpace(strings.TrimPrefix(cmd, "del connection "))
			err := deleteConnectionByName(connectionName)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Error: %v[white]\n", err)))
			} else {
				logDisplay.Write([]byte(fmt.Sprintf("[green]Connection '%s' deleted successfully[white]\n", connectionName)))

				// Refresh connections list
				connections, err := GetConnections()
				if err != nil {
					logDisplay.Write([]byte(fmt.Sprintf("[red]Error fetching connections: %v[white]\n", err)))
				} else {
					kvDisplay.SetText(FormatConnectionsList(connections))
				}
			}
			return
		case cmd == "del all connections":
			err := deleteAllConnections()
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Error: %v[white]\n", err)))
			} else {
				logDisplay.Write([]byte("[green]All connections deleted successfully[white]\n"))

				// Refresh connections list (which will now be empty)
				connections, err := GetConnections()
				if err != nil {
					logDisplay.Write([]byte(fmt.Sprintf("[red]Error fetching connections: %v[white]\n", err)))
				} else {
					kvDisplay.SetText(FormatConnectionsList(connections))
				}
			}
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
