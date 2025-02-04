package windows

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Amrit02102004/RediCLI/utils"
	"github.com/gdamore/tcell/v2"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rivo/tview"
	"github.com/xuri/excelize/v2"
)

// Win1.go
func ExportData(filePath string, redis *utils.RedisConnection) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"Key", "Value", "TTL"}); err != nil {
		return fmt.Errorf("error writing CSV header: %v", err)
	}

	keys, err := redis.GetAllKeys()
	if err != nil {
		return fmt.Errorf("error getting keys: %v", err)
	}

	for _, key := range keys {
		value, err := redis.GetValue(key)
		if err != nil {
			value = ""
		}

		ttl, err := redis.GetTTL(key)
		ttlStr := "-1"
		if err == nil && ttl > 0 {
			ttlStr = fmt.Sprintf("%.0f", ttl.Seconds())
		}

		if err := writer.Write([]string{key, value, ttlStr}); err != nil {
			return fmt.Errorf("error writing CSV row: %v", err)
		}
	}

	return nil
}

func ImportData(filePath string, redis *utils.RedisConnection) error {
	ext := filepath.Ext(filePath)
	var records [][]string

	if ext == ".csv" {
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("error opening CSV file: %v", err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err = reader.ReadAll()
		if err != nil {
			return fmt.Errorf("error reading CSV: %v", err)
		}
	} else if ext == ".xlsx" {
		f, err := excelize.OpenFile(filePath)
		if err != nil {
			return fmt.Errorf("error opening XLSX file: %v", err)
		}
		defer f.Close()

		rows, err := f.GetRows(f.GetSheetList()[0])
		if err != nil {
			return fmt.Errorf("error reading XLSX: %v", err)
		}
		records = rows
	} else {
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	if len(records) > 0 && len(records[0]) >= 3 {
		for _, row := range records[1:] {
			if len(row) >= 3 {
				key := row[0]
				value := row[1]
				ttl, err := strconv.ParseFloat(row[2], 64)
				if err != nil {
					ttl = -1
				}

				if ttl > 0 {
					err = redis.SetKeyWithTTL(key, value, time.Duration(ttl)*time.Second)
				} else {
					err = redis.SetKeyWithTTL(key, value, 0)
				}

				if err != nil {
					return fmt.Errorf("error setting key %s: %v", key, err)
				}
			}
		}
	}
	return nil
}

// Win2.go

// Win3.go

func DisplayWelcomeMessage(kvDisplay *tview.TextView) {
	welcome := `
[::b]       Welcome to RediCLI v1.0       [-:-:-]
[::b]──────────────────────────────────────[-:-:-]
[::b]  Type 'help' to see available   [-:-:-]
[::b]  commands and their descriptions    [-:-:-]
`
	kvDisplay.SetTextAlign(tview.AlignCenter) // Center the text
	kvDisplay.SetText(welcome)
}

func Clear(kvDisplay *tview.TextView, logDisplay *tview.TextView, x int, y int) {
    if x == 1 {
        kvDisplay.Clear()
        kvDisplay.SetTextAlign(tview.AlignCenter)
        DisplayWelcomeMessage(kvDisplay)
    }
    if y == 1 {
        logDisplay.Clear()
    }
}

// DisplayHelp shows all available commands and their descriptions
func DisplayHelp(kvDisplay *tview.TextView) {
	helpText := `[yellow]RediCLI v1.0 - Available Commands[-:-:-]

[::b]Basic Commands:[-:-:-]
  • [green]get <key>[-:-:-]
    Retrieve the value of a specified key
    
  • [green]set <key> <value>[-:-:-]
    Set a key with the specified value
    
  • [green]del <key>[-:-:-]
    Delete a specified key
    
  • [green]keys <pattern>[-:-:-]
    Find all keys matching the given pattern
    
  • [green]ttl <key>[-:-:-]
    Get the time to live for a key in seconds
    
  • [green]expire <key> <seconds>[-:-:-]
    Set a key's time to live in seconds

[::b]Advanced Commands:[-:-:-]
  • [green]key filter set[-:-:-]
    Open form to set a key with TTL in milliseconds
    
  • [green]key filter update[-:-:-]
    Open form to update a key with KEEPTTL option
    
  • [green]flushall[-:-:-]
    Delete all keys (use with caution)

[::b]Data Management:[-:-:-]
  • [green]import[-:-:-]
    Import data from CSV/XLSX file
    
  • [green]export[-:-:-]
    Export data to CSV file

[::b]Help:[-:-:-]
  • [green]help[-:-:-]
    Display this help message

[yellow]Note: Use TAB key to autocomplete commands[-:-:-]`

	kvDisplay.SetText(helpText)
	kvDisplay.SetTextAlign(tview.AlignLeft)
}

func ImportForm(app *tview.Application, redis *utils.RedisConnection, kvDisplay *tview.TextView, logDisplay *tview.TextView, cmdFlex *tview.Flex, formContainer *tview.Flex, suggestionDisplay *tview.TextView, cmdInput *tview.InputField) tview.Primitive {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Import Data ")

	filePathInput := tview.NewInputField().
		SetLabel("File path (.csv/.xlsx): ").
		SetFieldWidth(40)
	form.AddFormItem(filePathInput)

	form.AddButton("Import", func() {
		filePath := filePathInput.GetText()
		err := ImportData(filePath, redis)
		if err != nil {
			logDisplay.Write([]byte(fmt.Sprintf("[red]Import Error: %v[white]\n", err)))
		} else {
			logDisplay.Write([]byte("[green]Data imported successfully[white]\n"))
		}
		RefreshData(logDisplay, kvDisplay, redis)

		// Reset the view
		formContainer.AddItem(form, 0, 1, true)
		cmdFlex.RemoveItem(kvDisplay)
		cmdFlex.RemoveItem(suggestionDisplay)
		cmdFlex.RemoveItem(cmdInput)
		cmdFlex.RemoveItem(formContainer)
		cmdFlex.AddItem(kvDisplay, 0, 1, false)
		cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
		cmdFlex.AddItem(cmdInput, 1, 0, true)
	})

	form.AddButton("Cancel", func() {
		formContainer.AddItem(form, 0, 1, true)
		cmdFlex.RemoveItem(kvDisplay)
		cmdFlex.RemoveItem(suggestionDisplay)
		cmdFlex.RemoveItem(cmdInput)
		cmdFlex.RemoveItem(formContainer)
		cmdFlex.AddItem(kvDisplay, 0, 1, false)
		cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
		cmdFlex.AddItem(cmdInput, 1, 0, true)
	})

	form.SetBorderPadding(1, 1, 1, 1)
	return form
}

func ExportForm(app *tview.Application, redis *utils.RedisConnection, kvDisplay *tview.TextView, logDisplay *tview.TextView, cmdFlex *tview.Flex, formContainer *tview.Flex, suggestionDisplay *tview.TextView, cmdInput *tview.InputField) tview.Primitive {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Export Data ")

	filePathInput := tview.NewInputField().
		SetLabel("Export path (.csv): ").
		SetFieldWidth(40)
	form.AddFormItem(filePathInput)

	form.AddButton("Export", func() {
		filePath := filePathInput.GetText()
		if !strings.HasSuffix(filePath, ".csv") {
			filePath += ".csv"
		}
		err := ExportData(filePath, redis)
		if err != nil {
			logDisplay.Write([]byte(fmt.Sprintf("[red]Export Error: %v[white]\n", err)))
		} else {
			logDisplay.Write([]byte("[green]Data exported successfully[white]\n"))
		}
		RefreshData(logDisplay, kvDisplay, redis)

		// Reset the view
		formContainer.AddItem(form, 0, 1, true)
		cmdFlex.RemoveItem(kvDisplay)
		cmdFlex.RemoveItem(suggestionDisplay)
		cmdFlex.RemoveItem(cmdInput)
		cmdFlex.RemoveItem(formContainer)
		cmdFlex.AddItem(kvDisplay, 0, 1, false)
		cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
		cmdFlex.AddItem(cmdInput, 1, 0, true)
	})

	form.AddButton("Cancel", func() {
		formContainer.AddItem(form, 0, 1, true)
		cmdFlex.RemoveItem(kvDisplay)
		cmdFlex.RemoveItem(suggestionDisplay)
		cmdFlex.RemoveItem(cmdInput)
		cmdFlex.RemoveItem(formContainer)
		cmdFlex.AddItem(kvDisplay, 0, 1, false)
		cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
		cmdFlex.AddItem(cmdInput, 1, 0, true)
	})

	form.SetBorderPadding(1, 1, 1, 1)
	return form
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

	// Slice to track focusable items in order
	var focusableItems []tview.Primitive

	// Populate focusable items
	focusableItems = append(focusableItems, keyInput, valueInput, ttlInput)

	// Custom input capture to handle tab navigation
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			// Find the current focused item
			currentIndex := -1
			for i, item := range focusableItems {
				if item.HasFocus() {
					currentIndex = i
					break
				}
			}

			// Move to the next item
			if currentIndex != -1 {
				nextIndex := (currentIndex + 1) % len(focusableItems)
				app.SetFocus(focusableItems[nextIndex])
			}
			return nil
		}
		return event
	})

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
		cmdFlex.RemoveItem(kvDisplay)
		cmdFlex.RemoveItem(suggestionDisplay)
		cmdFlex.RemoveItem(cmdInput)
		cmdFlex.RemoveItem(formContainer)
		cmdFlex.AddItem(kvDisplay, 0, 1, false)
		cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
		cmdFlex.AddItem(cmdInput, 1, 0, true)
	})

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

	// Slice to track focusable items in order
	var focusableItems []tview.Primitive

	// Populate focusable items
	focusableItems = append(focusableItems, keyInput, valueInput, keepTTLCheckbox)

	// Custom input capture to handle tab navigation
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			// Find the current focused item
			currentIndex := -1
			for i, item := range focusableItems {
				if item.HasFocus() {
					currentIndex = i
					break
				}
			}

			// Move to the next item
			if currentIndex != -1 {
				nextIndex := (currentIndex + 1) % len(focusableItems)
				app.SetFocus(focusableItems[nextIndex])
			}
			return nil
		}
		return event
	})

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
		cmdFlex.RemoveItem(kvDisplay)
		cmdFlex.RemoveItem(suggestionDisplay)
		cmdFlex.RemoveItem(cmdInput)
		cmdFlex.RemoveItem(formContainer)
		cmdFlex.AddItem(kvDisplay, 0, 1, false)
		cmdFlex.AddItem(suggestionDisplay, 3, 0, false)
		cmdFlex.AddItem(cmdInput, 1, 0, true)
	})

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
