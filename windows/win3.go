package windows

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Amrit02102004/RediCLI/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Enhanced command suggestions with more metadata
type EnhancedCommandSuggestion struct {
	command     string
	description string
	category    string
}

// Expanded command suggestions with new features
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

func KeyFilterSetForm(app *tview.Application, redis *utils.RedisConnection, logDisplay *tview.TextView, kvDisplay *tview.TextView, flex *tview.Flex) *tview.Flex {
	// Create form
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

	// Create a modal-like overlay
	overlay := tview.NewFlex().SetDirection(tview.FlexRow)
	overlay.SetBackgroundColor(tcell.ColorBlack)

	// Center the form
	centeredForm := tview.NewFlex().SetDirection(tview.FlexColumn)
	centeredForm.AddItem(nil, 0, 1, false)
	centeredForm.AddItem(form, 40, 1, true)
	centeredForm.AddItem(nil, 0, 1, false)

	overlay.AddItem(nil, 0, 1, false)
	overlay.AddItem(centeredForm, 10, 1, true)
	overlay.AddItem(nil, 0, 1, false)

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
			app.SetRoot(flex, true)
		}
	})

	form.AddButton("Cancel", func() {
		app.SetRoot(flex, true)
	})

	return overlay
}

func KeyFilterUpdateForm(app *tview.Application, redis *utils.RedisConnection, logDisplay *tview.TextView, kvDisplay *tview.TextView, flex *tview.Flex) *tview.Flex {
	// Create form
	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Key Filter Update ")

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

	// Create a modal-like overlay
	overlay := tview.NewFlex().SetDirection(tview.FlexRow)
	overlay.SetBackgroundColor(tcell.ColorBlack)

	// Center the form
	centeredForm := tview.NewFlex().SetDirection(tview.FlexColumn)
	centeredForm.AddItem(nil, 0, 1, false)
	centeredForm.AddItem(form, 40, 1, true)
	centeredForm.AddItem(nil, 0, 1, false)

	overlay.AddItem(nil, 0, 1, false)
	overlay.AddItem(centeredForm, 10, 1, true)
	overlay.AddItem(nil, 0, 1, false)

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
			app.SetRoot(flex, true)
		}
	})

	form.AddButton("Cancel", func() {
		app.SetRoot(flex, true)
	})

	return overlay
}

// Enhanced suggestion filtering with fuzzy search
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
		SetTextColor(tcell.ColorGray)
	
	// Create key-value display
	kvDisplay := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	kvDisplay.SetBorder(true).SetTitle(" Redis Data ")
	
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
		
		// Store the original content
		// originalContent := kvDisplay.GetText(true)
		
		// Modify this part to correctly handle special commands
		switch {
		case cmd == "key filter set":
			form := KeyFilterSetForm(app, redis, logDisplay, kvDisplay, mainFlex)
			app.SetRoot(form, true)
			return
		case cmd == "key filter update":
			form := KeyFilterUpdateForm(app, redis, logDisplay, kvDisplay, mainFlex)
			app.SetRoot(form, true)
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
	cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
	cmdFlex.AddItem(cmdInput, 1, 0, true)

	return cmdFlex, kvDisplay, cmdInput, mainFlex
}

