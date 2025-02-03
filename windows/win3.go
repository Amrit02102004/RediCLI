package windows

import (
	"fmt"
	"strings"

	"github.com/Amrit02102004/RediCLI/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type CommandSuggestion struct {
    command     string
    description string
}

var commandSuggestions = []CommandSuggestion{
    {"get", "Retrieve the value of a key"},
    {"set", "Set the string value of a key"},
    {"del", "Delete a key"},
    {"keys", "Find all keys matching a pattern"},
    {"ttl", "Get the time to live for a key"},
    {"expire", "Set a key's time to live in seconds"},
}

func Win3(app *tview.Application, logDisplay *tview.TextView, redis *utils.RedisConnection) (*tview.Flex, *tview.TextView, *tview.InputField) {
    cmdFlex := tview.NewFlex().SetDirection(tview.FlexRow)
    
    // Create key-value display
    kvDisplay := tview.NewTextView().
        SetDynamicColors(true).
        SetChangedFunc(func() {
            app.Draw()
        })
    kvDisplay.SetBorder(true).SetTitle(" Redis Data ")
    
    // Create suggestion display
    suggestionDisplay := tview.NewTextView()
    suggestionDisplay.
        SetDynamicColors(true).
        SetTextColor(tcell.ColorGray).
        SetBackgroundColor(tcell.ColorDefault)
    suggestionDisplay.SetTextAlign(tview.AlignCenter)
    
    // Create command input field
    cmdInput := tview.NewInputField().
        SetLabel("> ").
        SetFieldWidth(0)
    
    currentSuggestionIndex := 0
    var currentSuggestions []CommandSuggestion

    cmdInput.SetChangedFunc(func(text string) {
        currentSuggestions = filterSuggestions(text)
        if len(currentSuggestions) > 0 {
            suggestionText := ""
            for _, sugg := range currentSuggestions {
                suggestionText += fmt.Sprintf("[gray]%s[white] - %s\n", sugg.command, sugg.description)
            }
            suggestionDisplay.Clear()
            fmt.Fprintf(suggestionDisplay, suggestionText)
        } else {
            suggestionDisplay.Clear()
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
    
    cmdInput.SetDoneFunc(func(key tcell.Key) {
        if key != tcell.KeyEnter {
            return
        }
        
        cmd := strings.TrimSpace(cmdInput.GetText())
        if cmd == "" {
            return
        }
        
        logDisplay.Write([]byte(fmt.Sprintf("> %s\n", cmd)))
        
        result, err := redis.ExecuteCommand(cmd)
        if err != nil {
            logDisplay.Write([]byte(fmt.Sprintf("[red]Error:[white] %v\n", err)))
        } else {
            logDisplay.Write([]byte(fmt.Sprintf("[green]Result:[white] %v\n", result)))
        }
        
        cmdInput.SetText("")
        RefreshData(logDisplay, kvDisplay, redis)  
    })
    
    cmdFlex.AddItem(kvDisplay, 0, 3, false).
        AddItem(suggestionDisplay, 1, 0, false). 
        AddItem(cmdInput, 1,1,false)
    return cmdFlex, kvDisplay, cmdInput
}

func filterSuggestions(input string) []CommandSuggestion {
    input = strings.ToLower(strings.TrimSpace(input))
    if input == "" {
        return nil
    }

    var matches []CommandSuggestion
    for _, sugg := range commandSuggestions {
        if strings.HasPrefix(sugg.command, input) {
            matches = append(matches, sugg)
        }
    }
    return matches
}