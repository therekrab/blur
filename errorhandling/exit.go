package errorhandling

import (
	"os"
	"therekrab/secrets/ui"
)

func Exit() {
    failed := hadError()
    if failed {
        ui.Err("Exiting with failure status.")
        ui.Cleanup()
        os.Exit(1)
    } else {
        ui.Out("Exiting regularly.")
        ui.Cleanup()
        os.Exit(0)
    }
}
