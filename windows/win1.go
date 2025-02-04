package windows

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/Amrit02102004/RediCLI/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/xuri/excelize/v2"
)

type KeyData struct {
	key    string
	value  string
	ttl    string
	memory int64
}

func ExportData(filePath string , redis *utils.RedisConnection) error {
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

	func 	ImportData(filePath string , redis *utils.RedisConnection) error {
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


func Win1(app *tview.Application, redis *utils.RedisConnection) *tview.Flex {
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
			ttlStr := "-1"
			if err == nil {
				if ttl < 0 {
					ttlStr = "-1"
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
			table.SetCell(rowIndex, 3, tview.NewTableCell(fmt.Sprintf("%d B", data.memory)).SetExpansion(1))
		}
	}


	

	buttons := tview.NewFlex().SetDirection(tview.FlexColumn)

	go func() {
		for {
			app.QueueUpdateDraw(refreshTableData)
			time.Sleep(1 * time.Second)
		}
	}()

	refreshTableData()

	mainFlex.AddItem(table, 0, 9, true)
	mainFlex.AddItem(buttons, 3, 1, false)
	return mainFlex
}
