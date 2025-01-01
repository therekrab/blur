package ui

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type userInterface struct {
    mu sync.Mutex
    active bool
    app *tview.Application
    flex *tview.Flex
    flexOuter *tview.Flex
    input *tview.InputField
    output *tview.TextView
    errorReport *tview.TextView
    inputChan chan string
}

var ui *userInterface
var once sync.Once

func configStyle() {
    tview.Styles.PrimitiveBackgroundColor = tcell.NewHexColor(0x282e2e)
    tview.Styles.ContrastBackgroundColor = tcell.NewHexColor(0x5a6c6c)
    tview.Styles.SecondaryTextColor = tcell.NewHexColor(0xabc8be)
}

func runUI() (err error) {
    err = ui.app.SetRoot(ui.flexOuter, true).Run()
    ui.mu.Lock()
    ui.active = false
    ui.mu.Unlock()
    return
}

func Init() {
    ui = &userInterface{}
    ui.inputChan = make(chan string, 1)
    ui.mu.Lock()
    defer ui.mu.Unlock()
    configStyle()
    // Setup the application
    ui.app = tview.NewApplication().
        SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
            if event.Key() == tcell.KeyCtrlC {
                return nil
            }
            return event
        })
    // Output: TextView
    ui.output = tview.NewTextView().
        SetDynamicColors(true).
        SetChangedFunc(func() {
            ui.output.ScrollToEnd()
            ui.app.Draw()
        })
    ui.output.SetBorder(true).
        SetBorderColor(tview.Styles.SecondaryTextColor).
        SetBorderPadding(1, 1, 2, 2).
        SetTitle("  Messages  ")
    // Input: InputField
    ui.input = tview.NewInputField().
        SetDoneFunc(func(key tcell.Key) {
            if key == tcell.KeyEsc {
                ui.app.Stop()
                ui.inputChan <- ""
                ui.active = false
            }
            if key == tcell.KeyEnter {
                ui.inputChan <- ui.input.GetText()
                ui.input.SetText("")
                ui.app.SetFocus(ui.input)
            }
        })
    // Error reporting: TextView
    ui.errorReport = tview.NewTextView().
        SetDynamicColors(true).
        SetTextColor(tcell.NewHexColor(0xe45050))
    ui.errorReport.SetChangedFunc(func() {
        ui.errorReport.ScrollToEnd()
        ui.app.Draw()
    })
    // Layout: Flex
    ui.flex = tview.NewFlex().
        SetDirection(tview.FlexRow).
        AddItem(tview.NewBox(), 0, 1, false).
        AddItem(ui.output, 0, 16, false).
        AddItem(ui.input, 0, 2, true).
        AddItem(ui.errorReport, 0, 1, false)
    ui.flexOuter = tview.NewFlex().
        SetDirection(tview.FlexColumn).
        AddItem(tview.NewBox(), 0, 1, false).
        AddItem(ui.flex, 0, 8, true).
        AddItem(tview.NewBox(), 0, 1, false)
    ui.active = true
}

func Run(done chan error) {
    err := runUI()
    done <- err
}

func Log(format string, a... any) {
    file, err := os.OpenFile(
        "blur.log",
        os.O_APPEND|os.O_CREATE|os.O_WRONLY,
        0644,
    )
    if err != nil {
        Err("%s", err.Error())
    }
    defer file.Close()
    // save the line
    fmt.Fprintf(file, format, a...)

}

func Out(format string, a... any) {
    if ui == nil || !ui.active {
        fmt.Printf(format, a...)
        return
    }
    ui.mu.Lock()
    defer ui.mu.Unlock()
    original := fmt.Sprintf(format, a...)
    safe := tview.Escape(original)
    fmt.Fprintf(ui.output, "%s", safe)
}

func OutBold(format string, a... any) {
    if ui == nil || !ui.active {
        fmt.Printf(format, a...)
        return
    }
    ui.mu.Lock()
    defer ui.mu.Unlock()
    original := fmt.Sprintf(format, a...)
    safe := tview.Escape(original)
    fmt.Fprintf(ui.output, "[::b]%s[::-]", safe)
}

func Err(format string, a... any) {
    if ui == nil || !ui.active {
        fmt.Fprintf(os.Stderr, format, a...)
        return
    }
    ui.mu.Lock()
    defer ui.mu.Unlock()
    ui.errorReport.SetText(tview.Escape(fmt.Sprintf(format, a...)))
}

func ReadInput(prompt string) (str string, err error) {
    if ui == nil || !ui.active {
        err = fmt.Errorf("UI not active")
        return
    }
    ui.input.SetLabel(prompt + " >> ")
    text := <- ui.inputChan
    ui.input.SetLabel("")
    str = strings.TrimSpace(text)
    return
}


func Cleanup() {
    if ui != nil && ui.active {
        ui.app.Stop()
    }
}
