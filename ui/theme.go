package ui

import (
	"bufio"
	"io"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UITheme struct {
    textColor tcell.Color
    altTextColor tcell.Color
    bgColor tcell.Color
    altBgColor tcell.Color
    errorColor tcell.Color
}

func (theme *UITheme) apply() {
    tview.Styles.PrimaryTextColor = theme.textColor
    tview.Styles.SecondaryTextColor = theme.altTextColor
    tview.Styles.PrimitiveBackgroundColor = theme.bgColor
    tview.Styles.ContrastBackgroundColor = theme.altBgColor
}

func LoadTheme(rdr bufio.Reader) (theme UITheme, err error) {
    for {
        // read a line
        var line string
        line, err = rdr.ReadString('\n')
        if err != nil {
            return
        }
        line = strings.TrimSpace(line)
        break
    }
    return
}

func DefaultTheme() UITheme {
    return UITheme{
        textColor: tcell.ColorWhite,
        altTextColor: tcell.NewHexColor(0xabc8be),
        bgColor: tcell.NewHexColor(0x282e2e),
        altBgColor: tcell.NewHexColor(0x516c6c),
        errorColor: tcell.ColorRed,
    }
}

func NoColorTheme() UITheme {
    return UITheme{
        textColor: tcell.ColorWhite,
        altTextColor: tcell.ColorWhiteSmoke,
        bgColor: tcell.ColorDarkGray,
        altBgColor: tcell.ColorGray,
        errorColor: tcell.ColorWhite,
    }
}
