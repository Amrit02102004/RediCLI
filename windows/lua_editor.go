package windows

import (
    "fmt"
    "strings"
    "github.com/gdamore/tcell/v2"
    "github.com/rivo/tview"
    "github.com/Amrit02102004/RediCLI/utils"
    lua "github.com/yuin/gopher-lua"
)

type LuaEditor struct {
    *tview.Flex
    textArea    *tview.TextArea
    outputArea  *tview.TextView
    redis       *utils.RedisConnection
    app         *tview.Application
    cmdFlex     *tview.Flex
    mainDisplay *tview.TextView
    L           *lua.LState
}

func createRedisModule(redis *utils.RedisConnection) *lua.LTable {
    L := lua.NewState()
    defer L.Close()
    
    mod := L.NewTable()
    
    // Helper to convert Go slice to Lua table
    sliceToTable := func(slice []string) *lua.LTable {
        tab := L.NewTable()
        for i, v := range slice {
            tab.RawSetInt(i+1, lua.LString(v))
        }
        return tab
    }

    // Add Redis commands as Lua functions
    mod.RawSetString("call", L.NewFunction(func(L *lua.LState) int {
        cmd := L.ToString(1)
        nargs := L.GetTop() - 1
        args := make([]string, nargs)
        
        for i := 0; i < nargs; i++ {
            args[i] = L.ToString(i + 2)
        }
        
        fullCmd := strings.Join(append([]string{cmd}, args...), " ")
        result, err := redis.ExecuteCommand(fullCmd)
        
        if err != nil {
            L.Push(lua.LNil)
            L.Push(lua.LString(err.Error()))
            return 2
        }
        
        // Convert result to Lua value
        switch v := result.(type) {
        case string:
            L.Push(lua.LString(v))
        case []string:
            L.Push(sliceToTable(v))
        case int64:
            L.Push(lua.LNumber(v))
        case nil:
            L.Push(lua.LNil)
        default:
            L.Push(lua.LString(fmt.Sprint(v)))
        }
        return 1
    }))
    
    return mod
}

func NewLuaEditor(app *tview.Application, redis *utils.RedisConnection, cmdFlex *tview.Flex, mainDisplay *tview.TextView) *LuaEditor {
    editor := &LuaEditor{
        Flex:        tview.NewFlex().SetDirection(tview.FlexRow),
        textArea:    tview.NewTextArea(),
        outputArea:  tview.NewTextView(),
        redis:       redis,
        app:         app,
        cmdFlex:     cmdFlex,
        mainDisplay: mainDisplay,
        L:           lua.NewState(),
    }

    // Configure text editor
    editor.textArea.SetBorder(true).SetTitle(" Lua Script Editor [Ctrl+R: Run] [ESC: Exit] ")
//     testScript := `-- Example: Set and get keys
// local redis = require("redis")

// -- Set a key
// local ok = redis.call("SET", "test_key", "Hello from GopherLua!")
// if not ok then
//     return "Failed to set key"
// end

// -- Get the key
// local value = redis.call("GET", "test_key")
// if not value then
//     return "Failed to get key"
// end

// -- Get all keys
// local keys = redis.call("KEYS", "*")
// if not keys then
//     return "Failed to get keys"
// end

// -- Return results
// return {
//     test_key_value = value,
//     total_keys = #keys}`
//     editor.textArea.SetText(testScript, true)

    // Configure output area
    editor.outputArea.SetBorder(true).SetTitle(" Output ")
    editor.outputArea.SetDynamicColors(true)

    // Layout
    editor.AddItem(editor.textArea, 0, 3, true)
    editor.AddItem(editor.outputArea, 0, 2, false)

    // Initialize Lua state
    editor.L.PreloadModule("redis", func(L *lua.LState) int {
        L.Push(createRedisModule(redis))
        return 1
    })

    // Global key handlers
    editor.textArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        switch event.Key() {
        case tcell.KeyCtrlR:  
            editor.runScript()
            return nil
        case tcell.KeyEsc:
            editor.exit()
            return nil
        }
        return event
    })

    return editor
}

func (e *LuaEditor) runScript() {
    defer func() {
        if r := recover(); r != nil {
            e.outputArea.SetText(fmt.Sprintf("[red]Panic in Lua script: %v[white]", r))
        }
    }()

    script := e.textArea.GetText()
    if strings.TrimSpace(script) == "" {
        e.outputArea.SetText("[red]Error: Script is empty[white]")
        return
    }

    // Create a new Lua state for each execution
    L := lua.NewState()
    defer L.Close()

    // Preload Redis module
    L.PreloadModule("redis", func(L *lua.LState) int {
        L.Push(createRedisModule(e.redis))
        return 1
    })

    if err := L.DoString(script); err != nil {
        e.outputArea.SetText(fmt.Sprintf("[red]Error executing script: %v[white]", err))
        return
    }

    // Get the result from the top of the stack
    result := L.Get(-1)
    L.Pop(1)

    // Format the result
    var output strings.Builder
    output.WriteString("[green]Result:[white]\n")
    
    switch v := result.(type) {
    case *lua.LTable:
        output.WriteString("{\n")
        v.ForEach(func(key, value lua.LValue) {
            output.WriteString(fmt.Sprintf("  %v = %v\n", key, value))
        })
        output.WriteString("}")
    default:
        output.WriteString(fmt.Sprint(result))
    }

    e.outputArea.SetText(output.String())
}

func (e *LuaEditor) exit() {
    e.L.Close()
    e.cmdFlex.Clear()
    e.cmdFlex.AddItem(e.mainDisplay, 0, 1, false)
    e.app.SetRoot(e.cmdFlex, true)
}