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

func Win1(app *tview.Application, redis *utils.RedisConnection) *tview.Flex {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	table := tview.NewTable().
	SetBorders(true).
	SetFixed(1, 0).
	SetSeparator(tview.Borders.Vertical)
	

	table.SetTitle(" Redis Key Details")
	table.SetTitleAlign(tview.AlignLeft)
	table.SetBorder(true)

	headerCells := []string{"Key", "Value", "TTL", "Memory"}
	for i, header := range headerCells {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false).
			SetExpansion(1)
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
			ttlStr := "-"
			if err == nil {
				if ttl < 0 {
					ttlStr = "-"
				} else {
					ttlStr = fmt.Sprintf("%.0f", ttl.Seconds())
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

			const maxValueLength = 50
			if len(value) > maxValueLength {
				value = value[:maxValueLength] + "..."
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
			table.SetCell(rowIndex, 0, tview.NewTableCell(data.key).SetExpansion(1))
			table.SetCell(rowIndex, 1, tview.NewTableCell(data.value).SetExpansion(1))
			table.SetCell(rowIndex, 2, tview.NewTableCell(data.ttl).SetExpansion(1))
			table.SetCell(rowIndex, 3, tview.NewTableCell(fmt.Sprintf("%d", data.memory)).SetExpansion(1))
		}
	}

	go func() {
		for {
			app.QueueUpdateDraw(refreshTableData)
			time.Sleep(1 * time.Second)
		}
	}()

	refreshTableData()

	flex.AddItem(table, 0, 1, true)

	return flex
}
