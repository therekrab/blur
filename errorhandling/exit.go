package errorhandling

import (
	"os"
	"therekrab/secrets/ui"
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
