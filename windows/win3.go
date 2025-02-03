package windows

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Amrit02102004/RediCLI/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rivo/tview"
)

type EnhancedCommandSuggestion struct {
	command     string
	description string
	category    string
}

var enhancedCommandSuggestions = []EnhancedCommandSuggestion{
	{"key filter set", "Open key set form with TTL in milliseconds", "Advanced"},
	{"key filter update", "Open key update form with KEEPTTL option", "Advanced"},
	{"get", "Retrieve the value of a key", "Basic"},
	{"set", "Set the string value of a key", "Basic"},
	{"del", "Delete a key", "Basic"},
	{"keys", "Find all keys matching a pattern", "Basic"},
	{"ttl", "Get the time to live for a key", "Basic"},
	{"expire", "Set a key's time to live in seconds", "Basic"},
}

func KeyFilterSetForm(app *tview.Application, redis *utils.RedisConnection, logDisplay *tview.TextView, kvDisplay *tview.TextView, flex *tview.Flex, formContainer *tview.Flex, cmdFlex *tview.Flex, suggestionDisplay *tview.TextView, cmdInput *tview.InputField) tview.Primitive {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Key Filter Set ")

	keyInput := tview.NewInputField().
		SetLabel("Key: ").
		SetFieldWidth(20)
	form.AddFormItem(keyInput)

	valueInput := tview.NewInputField().
		SetLabel("Value: ").
		SetFieldWidth(20)
	form.AddFormItem(valueInput)

	ttlInput := tview.NewInputField().
		SetLabel("TTL (ms): ").
		SetFieldWidth(10).
		SetAcceptanceFunc(tview.InputFieldInteger)
	form.AddFormItem(ttlInput)

	form.AddButton("Set Key", func() {
		key := keyInput.GetText()
		value := valueInput.GetText()
		ttlStr := ttlInput.GetText()

		if key == "" || value == "" {
			logDisplay.Write([]byte("[red]Error: Key and Value are required![white]\n"))
			return
		}

		var ttl time.Duration
		if ttlStr != "" {
			ttlMs, err := strconv.ParseInt(ttlStr, 10, 64)
			if err != nil {
				logDisplay.Write([]byte("[red]Error: Invalid TTL value![white]\n"))
				return
			}
			ttl = time.Duration(ttlMs) * time.Millisecond
		}

		err := redis.SetKeyWithTTL(key, value, ttl)
		if err != nil {
			logDisplay.Write([]byte(fmt.Sprintf("[red]Error setting key: %v[white]\n", err)))
		} else {
			logDisplay.Write([]byte(fmt.Sprintf("[green]Key '%s' set successfully with TTL %v[white]\n", key, ttl)))
			RefreshData(logDisplay, kvDisplay, redis)
			formContainer.AddItem(form, 0, 1, true)
			// formContainer.SetTitle(" Key Filter Set Form ")
			cmdFlex.RemoveItem(kvDisplay)
            cmdFlex.RemoveItem(suggestionDisplay)
            cmdFlex.RemoveItem(cmdInput)
            cmdFlex.RemoveItem(formContainer)
			cmdFlex.AddItem(kvDisplay, 0, 1, false)
            cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
            cmdFlex.AddItem(cmdInput, 1, 0, true)
		}
	})

	form.AddButton("Cancel", func() {
        formContainer.AddItem(form, 0, 1, true)
			// formContainer.SetTitle(" Key Filter Set Form ")
			cmdFlex.RemoveItem(kvDisplay)
            cmdFlex.RemoveItem(suggestionDisplay)
            cmdFlex.RemoveItem(cmdInput)
            cmdFlex.RemoveItem(formContainer)
			cmdFlex.AddItem(kvDisplay, 0, 1, false)
            cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
            cmdFlex.AddItem(cmdInput, 1, 0, true)
	})

	// Set a fixed size for the form
	form.SetBorderPadding(1, 1, 1, 1)
	return form
}

func KeyFilterUpdateForm(app *tview.Application, redis *utils.RedisConnection, logDisplay *tview.TextView, kvDisplay *tview.TextView, flex *tview.Flex, formContainer *tview.Flex, cmdFlex *tview.Flex, suggestionDisplay *tview.TextView, cmdInput *tview.InputField) tview.Primitive {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Key Update ")

	keyInput := tview.NewInputField().
		SetLabel("Key: ").
		SetFieldWidth(20)
	form.AddFormItem(keyInput)

	valueInput := tview.NewInputField().
		SetLabel("Value: ").
		SetFieldWidth(20)
	form.AddFormItem(valueInput)

	keepTTLCheckbox := tview.NewCheckbox().
		SetLabel("Keep TTL")
	form.AddFormItem(keepTTLCheckbox)

	form.AddButton("Update Key", func() {
		key := keyInput.GetText()
		value := valueInput.GetText()
		keepTTL := keepTTLCheckbox.IsChecked()

		if key == "" || value == "" {
			logDisplay.Write([]byte("[red]Error: Key and Value are required![white]\n"))
			return
		}

		exists, err := redis.KeyExists(key)
		if err != nil {
			logDisplay.Write([]byte(fmt.Sprintf("[red]Error checking key: %v[white]\n", err)))
			return
		}

		if !exists {
			logDisplay.Write([]byte("[red]Error: Key does not exist![white]\n"))
			return
		}

		err = redis.UpdateKey(key, value, keepTTL)
		if err != nil {
			logDisplay.Write([]byte(fmt.Sprintf("[red]Error updating key: %v[white]\n", err)))
		} else {
			logDisplay.Write([]byte(fmt.Sprintf("[green]Key '%s' updated successfully[white]\n", key)))
			RefreshData(logDisplay, kvDisplay, redis)
			formContainer.AddItem(form, 0, 1, true)
			// formContainer.SetTitle(" Key Filter Set Form ")
			cmdFlex.RemoveItem(kvDisplay)
            cmdFlex.RemoveItem(suggestionDisplay)
            cmdFlex.RemoveItem(cmdInput)
            cmdFlex.RemoveItem(formContainer)
			cmdFlex.AddItem(kvDisplay, 0, 1, false)
            cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
            cmdFlex.AddItem(cmdInput, 1, 0, true)
		}
	})

	form.AddButton("Cancel", func() {
        formContainer.AddItem(form, 0, 1, true)
			// formContainer.SetTitle(" Key Filter Set Form ")
			cmdFlex.RemoveItem(kvDisplay)
            cmdFlex.RemoveItem(suggestionDisplay)
            cmdFlex.RemoveItem(cmdInput)
            cmdFlex.RemoveItem(formContainer)
			cmdFlex.AddItem(kvDisplay, 0, 1, false)
            cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
            cmdFlex.AddItem(cmdInput, 1, 0, true)
	})

	// Set a fixed size for the form
	form.SetBorderPadding(1, 1, 1, 1)
	return form
}

func enhancedFilterSuggestions(input string) []EnhancedCommandSuggestion {
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return nil
	}

	var matches []EnhancedCommandSuggestion
	for _, sugg := range enhancedCommandSuggestions {
		if fuzzy.Match(input, sugg.command) {
			matches = append(matches, sugg)
		}
	}
	return matches
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
	kvDisplay.SetBorder(true).SetTitle(" Redis Data ")
	
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