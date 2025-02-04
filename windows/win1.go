package windows

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/Amrit02102004/RediCLI/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type KeyData struct {
	key    string
	value  string
	ttl    string
	memory int64
}

func Win1(app *tview.Application, redis *utils.RedisConnection, kvDisplay *tview.TextView) *tview.Flex {
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	table := tview.NewTable().
		SetBorders(true).
		SetFixed(1, 0).
		SetSeparator(tview.Borders.Vertical)

	headerCells := []string{"Key", "Value", "TTL", "Memory"}
	for i, header := range headerCells {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false)

		switch i {
		case 0: // Key
			cell.SetMaxWidth(15).SetExpansion(1)
		case 1: // Value
			cell.SetMaxWidth(30).SetExpansion(2)
		case 2: // TTL
			cell.SetMaxWidth(10).SetExpansion(1)
		case 3: // Memory
			cell.SetMaxWidth(10).SetExpansion(1)
		}

		table.SetCell(0, i, cell)
	}

	refreshTableData := func() {
		for i := table.GetRowCount() - 1; i > 0; i-- {
			table.RemoveRow(i)
		}

		if !redis.IsConnected() {
			return
		}

		keys, err := redis.GetAllKeys()
		if err != nil {
			table.SetCell(1, 0,
				tview.NewTableCell(fmt.Sprintf("Error: %v", err)).
					SetTextColor(tcell.ColorRed))
			return
		}

		allData := make([]KeyData, 0, len(keys))

		for _, key := range keys {
			value, err := redis.GetValue(key)
			if err != nil {
				value = fmt.Sprintf("Error: %v", err)
			}

			ttl, err := redis.GetTTL(key)
			ttlStr := "-1"
			if err == nil {
				if ttl < 0 {
					ttlStr = "-1"
				} else {
					ttlStr = fmt.Sprintf("%.0f s", ttl.Seconds())
				}
			}

			memoryCmd := fmt.Sprintf("memory usage %s", key)
			memoryResult, err := redis.ExecuteCommand(memoryCmd)
			var memoryBytes int64 = 0
			if err == nil {
				if memInt, err := strconv.ParseInt(fmt.Sprintf("%v", memoryResult), 10, 64); err == nil {
					memoryBytes = memInt
				}
			}

			// Truncate value for table display
			const maxValueLength = 30
			displayValue := value
			if len(displayValue) > maxValueLength {
				displayValue = displayValue[:maxValueLength] + "..."
			}

			allData = append(allData, KeyData{
				key:    key,
				value:  value,
				ttl:    ttlStr,
				memory: memoryBytes,
			})
		}

		sort.Slice(allData, func(i, j int) bool {
			return allData[i].memory < allData[j].memory
		})

		for i, data := range allData {
			rowIndex := i + 1
			table.SetCell(rowIndex, 0,
				tview.NewTableCell(data.key).
					SetMaxWidth(15).
					SetExpansion(1))

			table.SetCell(rowIndex, 1,
				tview.NewTableCell(data.value).
					SetMaxWidth(4).
					SetExpansion(2))

			table.SetCell(rowIndex, 2,
				tview.NewTableCell(data.ttl).
					SetMaxWidth(10).
					SetExpansion(1))

			table.SetCell(rowIndex, 3,
				tview.NewTableCell(fmt.Sprintf("%d B", data.memory)).
					SetMaxWidth(10).
					SetExpansion(1))
		}
	}

	// Modify selection changed function to handle long keys
	table.SetSelectionChangedFunc(func(row, column int) {
		if row > 0 && row < table.GetRowCount() {
			keyCell := table.GetCell(row, 0)
			if keyCell != nil {
				key := keyCell.Text
				fullValue, err := redis.GetValue(key)
				ttl, _ := redis.GetTTL(key)

				// Create a more detailed and formatted display
				displayText := fmt.Sprintf("[yellow]Key Details:[white]\n")
				displayText += fmt.Sprintf("Full Key: [green]%s[white]\n\n", key)

				displayText += fmt.Sprintf("[yellow]Value:[white]\n")
				if err != nil {
					displayText += fmt.Sprintf("[red]Error retrieving value:[white] %v\n", err)
				} else {
					// Show full value with better formatting
					displayText += fmt.Sprintf("%s\n\n", fullValue)
				}

				if ttl >= 0 {
					displayText += fmt.Sprintf("[yellow]TTL:[white] %.0f seconds\n", ttl.Seconds())
				} else {
					displayText += "[yellow]TTL:[white] Persistent (no expiration)\n"
				}

				kvDisplay.SetText(displayText)
			}
		}
	})

	go func() {
		for {
			app.QueueUpdateDraw(refreshTableData)
			time.Sleep(1 * time.Second)
		}
	}()

	refreshTableData()

	mainFlex.AddItem(table, 0, 10, true)

	return mainFlex
}
