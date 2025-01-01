package errorhandling

import (
	"os"
    "github.com/therekrab/blur/ui"
)

func Exit() {
    // Regular system exits
    failed := hadError()
    if failed {
        ui.Cleanup()
        os.Exit(1)
    } else {
        ui.Cleanup()
        os.Exit(0)
    }
}
