package main

import (
	"flag"
	"os"
	"strconv"
	"therekrab/secrets/client"
	"therekrab/secrets/errorhandling"
	"therekrab/secrets/server"
	"therekrab/secrets/ui"
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
    var sessionIDHex string
    var ident string
    var sessionKey string
    // Determine functionality
    if *serverFlag {
        server.RunServer(*port)
    } else {
        err := ui.ReadInput(&sessionKey, "Session key: ")
        if err != nil {
            errorhandling.Report(err, true)
            errorhandling.Exit()
        }
        err = ui.ReadInput(&ident, "ident: ")
        if err != nil {
            errorhandling.Report(err, true)
            errorhandling.Exit()
        }
        if *newFlag {
            doNew(*addr, sessionKey, ident)
        } else {
            err = ui.ReadInput(&sessionIDHex, "Session ID: ")
            if err != nil {
                errorhandling.Report(err, true)
                errorhandling.Exit()
            }
            sessionID, err := parseSessionID(sessionIDHex)
            if err != nil {
                errorhandling.Report(err, true)
                errorhandling.Exit()
            }
            doJoin(*addr, sessionID, sessionKey, ident)
        }
    }
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
        // it's been reported, we just need to exit
        os.Exit(1)
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
        os.Exit(1)
    }
}

func parseSessionID(sessionHex string) (sessionID uint16, err error) {
    sessionIDBig, err :=  strconv.ParseUint(sessionHex, 16, 16)
    sessionID = uint16(sessionIDBig)
    return
}
