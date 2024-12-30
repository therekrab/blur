package client

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"therekrab/secrets/errorhandling"
	"therekrab/secrets/message"
	"therekrab/secrets/sender"
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
    errorhandling.Exit()
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
        fmt.Print(">> ")
        rdr := bufio.NewReader(os.Stdin)
        line, err := rdr.ReadString('\n')
        // Check for ctrl-D EOF
        if err == io.EOF {
            return
        }
        if err != nil {
            errorhandling.Report(err, true)
            sender.SendError(client.conn)
            return
        }
        // Build and send the chat
        encryptedData, err := client.cfg.encrypt(line)
        if err != nil {
            errorhandling.Report(err, true)
            sender.SendError(client.conn)
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
    
    fmt.Println("=== ACTIVE USERS: ===")
    for _, reponseIdent := range reponseIdents {
        fmt.Printf("\t'%s'\n", reponseIdent)
    }
    fmt.Println("===== END USERS =====")
    return
}

func (client *Client) runOutputLoop() (err error) {
    defer client.Close()
    for client.isActive() {
        var msg message.Message
        msg, err = message.ReadMessage(client.conn)
        if !client.isActive() {
            return
        }
        if err != nil {
            errorhandling.Report(err, true)
            sender.SendError(client.conn)
            return
        }
        switch msg.MType() {
        case message.ERR:
            err = fmt.Errorf("received ERR msg")
            errorhandling.Report(err, true)
            return
        case message.IDENTR:
            err = client.identRoutine()
            if err != nil {
                errorhandling.Report(err, true)
                sender.SendError(client.conn)
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
                sender.SendError(client.conn)
                return
            }
            fmt.Printf("'%s' : %s\n", source, cht)
            continue
        case message.CHTE:
            var (
                source []byte
                chte []byte
            )
            source, chte, err = message.ParseCht(msg.Data())
            if err != nil {
                errorhandling.Report(err, true)
                sender.SendError(client.conn)
                return
            }
            var cht []byte
            cht, err = client.cfg.decrypt(chte)
            if err != nil {
                errorhandling.Report(err, true)
                sender.SendError(client.conn)
                return
            }
            fmt.Printf("'%s' : %s\n", source, string(cht))
            continue
        }
        // invalid type received
        err = fmt.Errorf("invalid type received: %d", msg.MType())
        errorhandling.Report(err, true)
        sender.SendError(client.conn)
    }
    return
}

func (client *Client) runLoop() {
    go client.runOutputLoop()
    client.runInputLoop()
    errorhandling.Exit()
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
            sender.SendError(client.conn)
            return
        }
        var response message.Message
        response, err = message.ReadMessage(client.conn)
        if err != nil {
            errorhandling.Report(err, true)
            sender.SendError(client.conn)
            return
        }
        switch response.MType() {
        case message.ACC:
            fmt.Printf("Joined session %x\n", client.cfg.sessionID)
            client.runLoop()
        case message.REJ:
            if response.Data()[0] == 0 {
                err = fmt.Errorf("invalid sessionID")
            } else {
                err = fmt.Errorf("incrorect credentials")
            }
            errorhandling.Report(err, true)
            return
        }
        err = fmt.Errorf("invalid response received from server")
        errorhandling.Report(err, true)
        return
    } else {
        err = sender.SendNewR(client.conn, client.cfg.HashedKey())
        if err != nil {
            errorhandling.Report(err, true)
            sender.SendError(client.conn)
            return
        }
        var response message.Message
        response, err = message.ReadMessage(client.conn)
        if err != nil {
            errorhandling.Report(err, true)
            sender.SendError(client.conn)
            return
        }
        switch response.MType() {
        case message.NEW:
            client.cfg.sessionID = binary.BigEndian.Uint16(response.Data())
            fmt.Printf("Created session %x\n", client.cfg.sessionID)
            client.runLoop()
            return
        }
        fmt.Println("did not receive NEW response")
    }
    return
}
