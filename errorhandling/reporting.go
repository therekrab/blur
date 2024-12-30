package errorhandling

import (
	"therekrab/secrets/ui"
)

func Report(err error, fatal bool) {
    // This is where we would add more error handling.
    ui.Err("err:\t%s\n", err)
    if fatal {
        fatalError()
    }
}

