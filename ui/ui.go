package ui

import (
	"fmt"
	"os"
    "bufio"
    "strings"
)

func Log(format string, a... any) {
    fmt.Printf(format, a...)
}

func Out(format string, a... any) {
    fmt.Printf(format, a...)
}

func Err(format string, a... any) {
    fmt.Fprintf(os.Stderr, format, a...)
}

func ReadInput(dst *string, prompt string) (err error) {
    fmt.Print(prompt)
    rdr := bufio.NewReader(os.Stdin)
    res, err := rdr.ReadString('\n')
    if err != nil {
        err = fmt.Errorf("could not read from stdin")
        return
    }
    *dst = strings.TrimSpace(res)
    if *dst == "" {
        err = fmt.Errorf("blank input not allowed")
    }
    return
}

func Cleanup() {}
