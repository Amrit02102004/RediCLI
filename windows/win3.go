// Create center command input and key-value display
package windows

import (
	"fmt"
	"strings"

	"github.com/Amrit02102004/RediCLI/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func Win3(app *tview.Application , logDisplay *tview.TextView, redis *utils.RedisConnection ) (*tview.Flex, *tview.TextView, *tview.InputField) {
    cmdFlex := tview.NewFlex().SetDirection(tview.FlexRow)
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
    
    

    return cmdFlex, kvDisplay, cmdInput
}