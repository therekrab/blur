package main

import (
	"flag"
	"strconv"
	"github.com/therekrab/blur/client"
	"github.com/therekrab/blur/errorhandling"
	"github.com/therekrab/blur/server"
	"github.com/therekrab/blur/ui"
)

func main() {
    port := flag.Uint("port", 4040,
        "(server mode) Port to run server on.",
    )
    serverFlag := flag.Bool("server", false,
        "Toggle server mode.",
    )
    newFlag := flag.Bool("new", false,
        "(client mode) Create a new session, rather than connecting to a prexisting one.",
    )
    addr := flag.String("addr", "127.0.0.1:4040",
        "(client mode) The server address to connect to.",
    )
    // Parse the flags
    flag.Parse()
    // Start the UI
    // Determine functionality
    if *serverFlag {
        server.RunServer(*port)
    } else {
        // Setup UI
        ui.Init()
        done := make(chan error)
        go ui.Run(done)
        sessionKey, err := ui.ReadInput("Session key: ")
        if err != nil {
            errorhandling.Exit()
        }
        ident, err := ui.ReadInput("ident: ")
        if err != nil {
            errorhandling.Exit()
        }
        if *newFlag {
            doNew(*addr, sessionKey, ident)
        } else {
            sessionIDHex, err := ui.ReadInput("Session ID: ")
            if err != nil {
                errorhandling.Exit()
            }
            sessionID, err := parseSessionID(sessionIDHex)
            if err != nil {
                errorhandling.Report(err, true)
            } else {
                doJoin(*addr, sessionID, sessionKey, ident)
            }
        }
        err = <- done
        if err != nil {
            errorhandling.Report(err, true)
        }
    }
    errorhandling.Exit()
}

func doNew(addr string, sessionKey string, ident string) {
    cfg, err := client.NewSessionConfig(sessionKey, ident)
    if err != nil {
        errorhandling.Report(err, true)
        return
    }
    c := client.NewClient(addr, cfg)
    err = c.Run(addr)
    if err != nil {
        errorhandling.Report(err, true)
    }
}

func doJoin(addr string, sessionID uint16, sessionKey string, ident string) {
    ui.Out("Attempting to join session %x\n", sessionID)
    cfg, err := client.JoinSessionConfig(sessionID, sessionKey, ident)
    if err != nil {
        errorhandling.Report(err, true)
        return
    }
    c := client.NewClient(addr, cfg)
    err = c.Run(addr)
    if err != nil {
        errorhandling.Report(err, true)
    }
}

func parseSessionID(sessionHex string) (sessionID uint16, err error) {
    sessionIDBig, err :=  strconv.ParseUint(sessionHex, 16, 16)
    sessionID = uint16(sessionIDBig)
    return
}
