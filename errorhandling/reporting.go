package errorhandling

import (
	"fmt"
	"os"
)

func Report(err error, fatal bool) {
    // This is where we would add more error handling.
    fmt.Fprintf(os.Stderr, "err:\t%s\n", err)
    if fatal {
        fatalError()
    }
}

