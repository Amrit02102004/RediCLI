package cmd

import (
	"time"

	"github.com/Amrit02102004/RediCLI/utils"
	"github.com/Amrit02102004/RediCLI/windows"
	"github.com/rivo/tview"
)

func Func() {
	app := tview.NewApplication()
	redis := utils.NewRedisConnection()

	flex := tview.NewFlex()
	logDisplay := windows.Win2(app)
	cmdFlex, kvDisplay, cmdInput := windows.Win3(app, logDisplay, redis)

	form := windows.ConnectionForm(app, logDisplay, redis, kvDisplay)
	flex.AddItem(form, 30, 1, true).
		AddItem(cmdFlex, 0, 2, false).
		AddItem(logDisplay, 30, 1, false)

	go func() {
		for {
			if redis.IsConnected() {
				app.QueueUpdateDraw(func() {
					flex.RemoveItem(form) 
					flex.RemoveItem(cmdFlex) 
					flex.RemoveItem(logDisplay)
					flex.AddItem(windows.Win1(app, redis), 40, 1, true).
                        AddItem(cmdFlex, 0, 2, false).
		                AddItem(logDisplay, 30, 1, false)
				})
				break
			}
			time.Sleep(1 * time.Second) 
		}
	}()

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
		windows.RefreshData(logDisplay, kvDisplay, redis)
	})

	cmdFlex.AddItem(kvDisplay, 0, 3, false).
		AddItem(cmdInput, 3, 1, true)

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
