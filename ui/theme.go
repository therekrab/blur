package ui

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var hexRegex *regexp.Regexp = regexp.MustCompile(
    `^#[\da-f]{6}$`,
)

type UITheme struct {
    text tcell.Color
    altText tcell.Color
    bg tcell.Color
    altBg tcell.Color
    error tcell.Color
}

type UIConfig struct {
    Text string
    AltText string
    Bg string
    AltBg string
    Error string
}

var instance *UITheme

func theme() *UITheme {
    if instance == nil {
        def := defaultTheme()
        instance = &def
    }
    return instance
}

func defaultTheme() UITheme {
    def, err := loadTheme("mono")
    if err != nil {
        // Use term default values
        def = UITheme{}
    }
    return def
}

func apply() {
    tview.Styles.PrimaryTextColor = theme().text
    tview.Styles.SecondaryTextColor = theme().altText
    tview.Styles.PrimitiveBackgroundColor = theme().bg
    tview.Styles.ContrastBackgroundColor = theme().altBg
}

func SetTheme(themeName string) (err error) {
    th, err := loadTheme(themeName)
    if err != nil {
        return
    }
    instance = &th
    return
}

func loadTheme(themeName string) (theme UITheme, err error) {
    if themeName == "" {
        theme = defaultTheme()
        return
    }
    // Check for the theme in the themes directory of the system
    home, ok := os.LookupEnv("HOME")
    if !ok {
        err = fmt.Errorf("Could not detect HOME directory")
        return
    }
    themeFilepath := fmt.Sprintf("%s/.blur/themes/%s.toml", home, themeName)
    var conf UIConfig
    if _, err = toml.DecodeFile(themeFilepath, &conf); err != nil {
        if os.IsNotExist(err) {
            err = fmt.Errorf("Invalid theme: %s", themeName)
        }
        return
    }
    theme, err = buildTheme(&conf)
    return
}

func buildTheme(conf *UIConfig) (theme UITheme, err error) {
    theme.text, err = readColor(conf.Text)
    if err != nil {
        return
    }
    theme.altText, err = readColor(conf.AltText)
    if err != nil {
        return
    }
    theme.bg, err = readColor(conf.Bg)
    if err != nil {
        return
    }
    theme.altBg, err = readColor(conf.AltBg)
    if err != nil {
        return
    }
    theme.error, err = readColor(conf.Error)
    return
}

func readColor(str string) (color tcell.Color, err error) {
    if res, ok := tcell.ColorNames[str]; ok {
        color = res
        return
    }
    if hexRegex.MatchString(str) {
        hex := strings.TrimPrefix(str, "#")
        var hexValue int64
        hexValue, err = strconv.ParseInt(hex, 16, 32)
        if err != nil {
            return
        }
        color = tcell.NewHexColor(int32(hexValue))
        return
    }
    return
}
