package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"therekrab/secrets/client"
	"therekrab/secrets/errorhandling"
	"therekrab/secrets/server"
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
        readInput(&sessionKey, "Session key: ")
        readInput(&ident, "ident: ")
        if *newFlag {
            doNew(*addr, sessionKey, ident)
        } else {
            readInput(&sessionIDHex, "Session ID: ")
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
    fmt.Printf("Attempting to join session %x\n", sessionID)
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

func readInput(dst *string, msg string) {
    fmt.Print(msg)
    rdr := bufio.NewReader(os.Stdin)
    res, err := rdr.ReadString('\n')
    if err != nil {
        err = fmt.Errorf("could not read from stdin")
        errorhandling.Report(err, true)
        errorhandling.Exit()
    }
    *dst = strings.TrimSpace(res)
    if *dst == "" {
        err = fmt.Errorf("blank input not allowed")
        errorhandling.Report(err, true)
        errorhandling.Exit()
    }
    return
}

func parseSessionID(sessionHex string) (sessionID uint16, err error) {
    sessionIDBig, err :=  strconv.ParseUint(sessionHex, 16, 16)
    sessionID = uint16(sessionIDBig)
    return
}
