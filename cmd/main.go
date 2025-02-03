package cmd

import (

	"github.com/Amrit02102004/RediCLI/utils"
	"github.com/Amrit02102004/RediCLI/windows"
	"github.com/rivo/tview"
)

func Func() {
    app := tview.NewApplication()
    redis := utils.NewRedisConnection()
    
    // Create the main flex container
    logDisplay := windows.Win2(app)
    cmdFlex, kvDisplay, cmdInput, flex := windows.Win3(app, logDisplay, redis)
    
    // Create Redis connection instance
    form := windows.ConnectionForm(app, logDisplay, redis, kvDisplay)
    
    // Add command input and key-value display to center flex
    cmdFlex.AddItem(kvDisplay, 0, 3, false).
        AddItem(cmdInput, 3, 1, true)
    
    // Create the layout with three panels
    flex.AddItem(form, 30, 1, true).
        AddItem(cmdFlex, 0, 2, false).
        AddItem(logDisplay, 30, 1, false)
    
    // Run the application
    if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
        panic(err)
    }
}