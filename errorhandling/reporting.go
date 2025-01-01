package errorhandling

import (
	"github.com/therekrab/blur/ui"
)

func Report(err error, fatal bool) {
    ui.Err("err:\t%s\n", err)
    if fatal {
        fatalError()
    }
}

func Log(err error, fatal bool) {
    ui.Log("[ ERR ] %s\n", err)
    Report(err, fatal)
}

