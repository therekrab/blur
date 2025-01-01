package client

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"github.com/therekrab/blur/errorhandling"
	"github.com/therekrab/blur/message"
	"github.com/therekrab/blur/sender"
	"github.com/therekrab/blur/ui"
)

type Client struct {
    mu sync.Mutex
    conn net.Conn
    active bool
    cfg ClientConfig
}

func NewClient(addr string, cfg ClientConfig) (client Client) {
    client.cfg = cfg
    return
}

func (client *Client) Close() {
    client.mu.Lock()
    defer client.mu.Unlock()
    if client.active == false {
        // we're already closed
        return
    }
    client.conn.Close()
    client.active = false
}

func (client *Client) isActive() bool {
    client.mu.Lock()
    defer client.mu.Unlock()
    return client.active
}

func (client *Client) runInputLoop() {
    defer client.Close()
    for client.isActive() {
        // get input from the user
        line, err := ui.ReadInput(string(client.cfg.ident))
        if err != nil {
            client.Close()
            errorhandling.Exit()
        }
        if line == ".exit" {
            // the client would like to leave
            client.Close()
            errorhandling.Exit()
        }
        if line == ".help" {
            ui.Out("==== HELP (Your eyes only) ====\n")
            ui.Out("\tType .help to see this message again.\n")
            ui.Out("\tType .exit to leave the chat.\n")
            ui.Out("\t<Esc> will also quit.\n")
            continue // noo dont send that
        }
        // Build and send the chat
        encryptedData, err := client.cfg.encrypt(line)
        if err != nil {
            errorhandling.Report(err, true)
            return
        }
        sender.SendChatE(client.conn, encryptedData)
    }
}

func (client *Client) identRoutine() (err error) {
    ident := client.cfg.ident
    idents := make([][]byte, 1)
    idents[0] = ident
    err = sender.SendIdent(client.conn, idents)
    if err != nil {
        return
    }
    // Now request identification!
    err = sender.SendIdentR(client.conn)
    if err != nil {
        return
    }
    response, err := message.ReadMessage(client.conn)
    if err != nil {
        return
    }
    if response.MType() != message.IDENT {
        err = fmt.Errorf(
            "invalid response to IDENTR received: %v",
            response.MType(),
        )
        return
    }
    reponseIdents, err := message.ParseIdent(response.Data())
    if err != nil {
        return
    }
    
    ui.OutBold("=== ACTIVE USERS: ===\n")
    for _, reponseIdent := range reponseIdents {
        ui.Out("\t'%s'\n", reponseIdent)
    }
    ui.OutBold("===== END USERS =====\n")
    return
}

func (client *Client) runOutputLoop() (err error) {
    defer client.Close()
    for client.isActive() {
        var msg message.Message
        msg, err = message.ReadMessage(client.conn)
        // The last line was blocking, so we may actually not be active anymore
        if !client.isActive() {
            return
        }
        if err != nil {
            errorhandling.Report(err, true)
            return
        }
        switch msg.MType() {
        case message.IDENTR:
            err = client.identRoutine()
            if err != nil {
                errorhandling.Report(err, true)
                return
            }
            continue
        case message.CHT:
            var (
                source []byte
                cht []byte
            )
            source, cht, err = message.ParseCht(msg.Data())
            if err != nil {
                errorhandling.Report(err, true)
                return
            }
            ui.Out("'%s' : %s\n", source, cht)
            continue
        case message.CHTE:
            var (
                source []byte
                chte []byte
            )
            source, chte, err = message.ParseCht(msg.Data())
            if err != nil {
                errorhandling.Report(err, true)
                return
            }
            var cht []byte
            cht, err = client.cfg.decrypt(chte)
            if err != nil {
                errorhandling.Report(err, true)
                return
            }
            ui.Out("'%s' : %s\n", source, string(cht))
            continue
        }
        // invalid type received
        err = fmt.Errorf("invalid type received: %d", msg.MType())
        errorhandling.Report(err, true)
    }
    return
}

func (client *Client) runLoop() {
    go client.runOutputLoop()
    client.runInputLoop()
}

func (client *Client) Run(addr string) (err error) {
    client.conn, err = net.Dial("tcp", addr)
    if err != nil {
        errorhandling.Report(err, true)
        return
    }
    client.active = true
    if client.cfg.join {
        // send a JOINR request
        err = sender.SendJoinR(
            client.conn,
            client.cfg.sessionID,
            client.cfg.HashedKey(),
        )
        if err != nil {
            errorhandling.Report(err, true)
            return
        }
        var response message.Message
        response, err = message.ReadMessage(client.conn)
        if err != nil {
            errorhandling.Report(err, true)
            return
        }
        switch response.MType() {
        case message.ACC:
            ui.Out("Joined session %x\n", client.cfg.sessionID)
            client.runLoop()
            return
        case message.REJ:
            if response.Data()[0] == 0 {
                err = fmt.Errorf("invalid sessionID")
            } else {
                err = fmt.Errorf("incrorect credentials")
            }
            errorhandling.Report(err, true)
            return
        }
        err = fmt.Errorf(
            "invalid response received from server (%d)",
            response.MType(),
        )
        errorhandling.Report(err, true)
    } else {
        err = sender.SendNewR(client.conn, client.cfg.HashedKey())
        if err != nil {
            errorhandling.Report(err, true)
            return
        }
        var response message.Message
        response, err = message.ReadMessage(client.conn)
        if err != nil {
            errorhandling.Report(err, true)
            return
        }
        if response.MType() == message.NEW {
            client.cfg.sessionID = binary.BigEndian.Uint16(response.Data())
            ui.Out("Created session %x\n", client.cfg.sessionID)
            client.runLoop()
        } else {
            err = fmt.Errorf("Failed creating new session")
            errorhandling.Report(err, true)
        }
    }
    return
}
