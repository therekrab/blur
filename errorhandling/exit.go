package errorhandling

import (
	"fmt"
	"os"
)

func Exit() {
    failed := hadError()
    if failed {
        fmt.Println("Exiting with failure status.")
        os.Exit(1)
    } else {
        fmt.Println("Exiting regularly.")
        os.Exit(0)
    }
}
