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

	logDisplay := windows.Win2(app)
	cmdFlex, kvDisplay, _, flex := windows.Win3(app, logDisplay, redis)

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


	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
