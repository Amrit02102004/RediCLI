package windows

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"os"

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
	{"quit", "Exit the RediCLI application", "Basic"},
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
	{"help", "Display help information and available commands", "Help"},
	{"clear all", "clear console and logs screen", "Basic"},
	{"clear logs", "clear logs screen", "Basic"},
	{"clear display", "clear display screen", "Basic"},
	{"add connection", "Open form to add and connect to a new Redis connection", "Connection Management"},
	{"view all connections", "List all saved Redis connections", "Connection Management"},
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
		case tcell.KeyEnter:
			// Clear suggestion display on enter
			suggestionDisplay.SetText("")
			// cmdInput.SetText("")
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

		// Assuming you have a `redis` package handling Redis operations

		if strings.HasPrefix(cmd, "get ") {
			// Extract the key name
			keyName := strings.TrimSpace(strings.TrimPrefix(cmd, "get"))

			// Check if key exists
			exists, err := redis.KeyExists(keyName)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Error checking key:[white] %v\n", err)))
				cmdInput.SetText("")
				return
			}

			if !exists {
				logDisplay.Write([]byte(fmt.Sprintf("[yellow]Key '%s' does not exist[white]\n", keyName)))
				kvDisplay.Clear()
				DisplayWelcomeMessage(kvDisplay)
				cmdInput.SetText("")
				return
			}

			// Get the key value
			value, err := redis.GetValue(keyName)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Error getting value:[white] %v\n", err)))
				cmdInput.SetText("")
				return
			}

			// Log key existence
			logDisplay.Write([]byte(fmt.Sprintf("[green]Key '%s' found[white]\n", keyName)))

			// Try pretty-printing JSON
			var prettyJSON string
			var formattedJSON map[string]interface{}

			if err := json.Unmarshal([]byte(value), &formattedJSON); err == nil {
				// Successfully parsed JSON, format it
				if prettyJSONBytes, err := json.MarshalIndent(formattedJSON, "", "  "); err == nil {
					prettyJSON = string(prettyJSONBytes)
				} else {
					// If marshalling fails, fall back to raw value
					prettyJSON = value
				}
			} else {
				// If unmarshalling fails, fall back to raw value
				prettyJSON = value
			}

			// Get the TTL
			ttl, err := redis.GetTTL(keyName)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Error getting TTL:[white] %v\n", err)))
				cmdInput.SetText("")
				return
			}

			// Clear previous display and show key details
			kvDisplay.Clear()

			// Format TTL display
			var ttlDisplay string
			switch {
			case ttl == -1:
				ttlDisplay = "No expiration"
			case ttl == -2:
				ttlDisplay = "Key does not exist"
			default:
				ttlDisplay = fmt.Sprintf("%v remaining", ttl.Round(time.Second))
			}

			// Display key details
			kvDisplay.SetText(fmt.Sprintf(
				"[green]Key Information:[white]\n\n"+
					"[yellow]Key Name:[white] %s\n\n"+
					"[yellow]Value:[white]\n%s\n\n"+
					"[yellow]Time to Live (TTL):[white] %s",
				keyName, prettyJSON, ttlDisplay,
			)).SetTextAlign(tview.AlignLeft)

			return
		}

		if strings.HasPrefix(cmd, "import .") {
			// Extract the file path
			filePath := strings.TrimSpace(strings.TrimPrefix(cmd, "import"))
			err := ImportData(filePath, redis)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Import Error: %v[white]\n", err)))
			} else {
				logDisplay.Write([]byte("[green]Data imported successfully[white]\n"))
			}
			RefreshData(logDisplay, kvDisplay, redis)
			return
		}
		
		if strings.HasPrefix(cmd, "export .") {
			// Extract the file path
			filePath := strings.TrimSpace(strings.TrimPrefix(cmd, "export"))
			err := ExportData(filePath, redis)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Import Error: %v[white]\n", err)))
			} else {
				logDisplay.Write([]byte("[green]Data Exported successfully[white]\n"))
			}
			RefreshData(logDisplay, kvDisplay, redis)
			return
		}

		switch {
		case cmd == "quit":
			// Clean up resources if needed
			if redis.IsConnected() {
				redis.Close()
			}
			app.Stop()
			os.Exit(0)
			return

		case cmd == "see analytics":
			go func() {
				err := redis.ServeAnalytics()
				if err != nil {
					logDisplay.Write([]byte(fmt.Sprintf("[red]Analytics Server Error:[white] %v\n", err)))
					cmdInput.SetText("")
				}
			}()
			logDisplay.Write([]byte("[green]Analytics server started on http://localhost:8080[white]\n"))
			cmdInput.SetText("")
			return
		case strings.HasPrefix(cmd, "select from"):
			condition, err := ParseQuery(cmd)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Query Error:[white] %v\n", err)))
				cmdInput.SetText("")
				return
			}
		
			// If a different connection is specified, connect to it
			if condition.ConnectionName != "" {
				config, err := FindConnectionByName(condition.ConnectionName)
				if err != nil {
					logDisplay.Write([]byte(fmt.Sprintf("[red]Connection Error:[white] %v\n", err)))
					cmdInput.SetText("")
					return
				}
		
				err = redis.Connect(config.Host, config.Port)
				if err != nil {
					logDisplay.Write([]byte(fmt.Sprintf("[red]Connection Error:[white] %v\n", err)))
					cmdInput.SetText("")
					return
				}
			}
		
			results, err := ExecuteQuery(redis, condition)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Query Error:[white] %v\n", err)))
				cmdInput.SetText("")
				return
			}
		
			// Display results
			var displayText strings.Builder
			displayText.WriteString("[green]Query Results:[white]\n\n")
			
			if len(results) == 0 {
				displayText.WriteString("No matching keys found.\n")
			} else {
				for key, value := range results {
					ttl, _ := redis.GetTTL(key)
					displayText.WriteString(fmt.Sprintf("[yellow]Key:[white] %s\n", key))
					displayText.WriteString(fmt.Sprintf("[yellow]Value:[white] %s\n", value))
					displayText.WriteString(fmt.Sprintf("[yellow]TTL:[white] %v\n\n", ttl))
				}
			}
		
			kvDisplay.Clear()
			kvDisplay.SetText(displayText.String()).SetTextAlign(tview.AlignLeft)
			cmdInput.SetText("")
			return
		case strings.HasPrefix(cmd, "update"):
			updateQuery, err := ParseUpdateQuery(cmd)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Update Query Error:[white] %v\n", err)))
				cmdInput.SetText("")
				return
			}
		
			// If a different connection is specified, connect to it
			if updateQuery.ConnectionName != "" {
				config, err := FindConnectionByName(updateQuery.ConnectionName)
				if err != nil {
					logDisplay.Write([]byte(fmt.Sprintf("[red]Connection Error:[white] %v\n", err)))
					cmdInput.SetText("")
					return
				}
		
				err = redis.Connect(config.Host, config.Port)
				if err != nil {
					logDisplay.Write([]byte(fmt.Sprintf("[red]Connection Error:[white] %v\n", err)))
					cmdInput.SetText("")
					return
				}
			}
		
			// Execute the update
			updatedCount, err := ExecuteUpdateQuery(redis, updateQuery)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Update Error:[white] %v\n", err)))
				cmdInput.SetText("")
				return
			}
		
			logDisplay.Write([]byte(fmt.Sprintf("[green]Successfully updated %d keys[white]\n", updatedCount)))
			RefreshData(logDisplay, kvDisplay, redis)
			cmdInput.SetText("")
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
			cmdInput.SetText("")
			return
		case cmd == "clear all":
			Clear(kvDisplay, logDisplay, 1, 1)
			cmdFlex.Clear()
			cmdInput.SetText("")
			cmdFlex.AddItem(kvDisplay, 0, 1, false)
			cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
			cmdFlex.AddItem(cmdInput, 1, 0, true)
			return
		case cmd == "clear logs":
			Clear(kvDisplay, logDisplay, 0, 1)
			cmdInput.SetText("")
			return
		case cmd == "clear display":
			Clear(kvDisplay, logDisplay, 1, 0)
			cmdInput.SetText("")
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
			cmdInput.SetText("")
			return

		case cmd == "key filter update":
			form := KeyFilterUpdateForm(app, redis, logDisplay, kvDisplay, mainFlex, formContainer, cmdFlex, suggestionDisplay, cmdInput)
			formContainer.Clear()
			cmdFlex.Clear()
			formContainer.AddItem(form, 0, 1, true)
			cmdFlex.RemoveItem(kvDisplay)
			cmdFlex.RemoveItem(suggestionDisplay)
			cmdFlex.RemoveItem(cmdInput)
			cmdInput.SetText("")

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
			cmdInput.SetText("")
			return

		case cmd == "export":
			form := ExportForm(app, redis, kvDisplay, logDisplay, cmdFlex, formContainer, suggestionDisplay, cmdInput)
			formContainer.Clear()
			cmdFlex.Clear()
			formContainer.AddItem(form, 0, 1, true)
			cmdFlex.AddItem(formContainer, 0, 1, false)
			cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
			cmdFlex.AddItem(cmdInput, 1, 0, true)
			cmdInput.SetText("")
			return

		case cmd == "help":
			cmdFlex.Clear()
			DisplayHelp(kvDisplay)
			cmdFlex.AddItem(kvDisplay, 0, 1, false)
			cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
			cmdFlex.AddItem(cmdInput, 1, 0, true)
			cmdInput.SetText("")
			return

		case strings.HasPrefix(cmd, "add connection"):
			form := ConnectionForm(app, logDisplay, redis, kvDisplay)
			formContainer.Clear()
			cmdFlex.Clear()
			formContainer.AddItem(form, 0, 1, true)
			cmdFlex.AddItem(formContainer, 0, 1, false)
			cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
			cmdFlex.AddItem(cmdInput, 1, 0, true)
			cmdInput.SetText("")
			return

		case cmd == "view all connections":
			cmdFlex.Clear()
			connections, err := GetConnections()
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Error fetching connections: %v[white]\n", err)))
			} else {
				kvDisplay.SetText(FormatConnectionsList(connections)).SetTextAlign(tview.AlignLeft)
			}
			cmdFlex.AddItem(kvDisplay, 0, 1, false)
			cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
			cmdFlex.AddItem(cmdInput, 1, 0, true)
			cmdInput.SetText("")
			return

		case strings.HasPrefix(cmd, "connect "):
			connectionName := strings.TrimSpace(strings.TrimPrefix(cmd, "connect"))
			config, err := FindConnectionByName(connectionName)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]%v[white]\n", err)))
				cmdInput.SetText("")
				return
			}

			err = redis.Connect(config.Host, config.Port)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Connection failed: %v[white]\n", err)))
				cmdInput.SetText("")
			} else {
				logDisplay.Write([]byte(fmt.Sprintf("[green]Connected to '%s' at %s:%s[white]\n",
					config.Name, config.Host, config.Port)))
				RefreshData(logDisplay, kvDisplay, redis)
				cmdInput.SetText("")
			}
			return
		case strings.HasPrefix(cmd, "del connection "):
			connectionName := strings.TrimSpace(strings.TrimPrefix(cmd, "del connection "))
			err := deleteConnectionByName(connectionName)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Error: %v[white]\n", err)))
				cmdInput.SetText("")
			} else {
				logDisplay.Write([]byte(fmt.Sprintf("[green]Connection '%s' deleted successfully[white]\n", connectionName)))
				cmdInput.SetText("")
				// Refresh connections list
				connections, err := GetConnections()
				if err != nil {
					logDisplay.Write([]byte(fmt.Sprintf("[red]Error fetching connections: %v[white]\n", err)))
				} else {
					kvDisplay.SetText(FormatConnectionsList(connections)).SetTextAlign(tview.AlignLeft)
				}
				cmdInput.SetText("")
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
					kvDisplay.SetText(FormatConnectionsList(connections)).SetTextAlign(tview.AlignLeft)
				}
			}
			cmdInput.SetText("")
			return

		default:
			result, err := redis.ExecuteCommand(cmd)
			if err != nil {
				logDisplay.Write([]byte(fmt.Sprintf("[red]Error:[white] %v\n", err)))
			} else {
				logDisplay.Write([]byte(fmt.Sprintf("[green]Result:[white] %v\n", result)))
			}
			cmdInput.SetText("")
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
