package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"github.com/therekrab/blur/errorhandling"
	"github.com/therekrab/blur/manager"
	"github.com/therekrab/blur/message"
	"github.com/therekrab/blur/sender"
	"github.com/therekrab/blur/ui"
)

func RunServer(port uint) (err error) {
    addr := fmt.Sprintf("0.0.0.0:%d", port)
    ln, err := net.Listen("tcp", addr)
    // handle any SIGINTS to gracefully shut down
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGKILL)
    active := true
    activeMu := new(sync.Mutex)
    go func() {
        <- sigs
        ui.Out("\nExiting...\n")
        ui.Log("[ SERVER ] Received exit signal, exiting\n")
        activeMu.Lock()
        active = false
        activeMu.Unlock()
        ln.Close()
        errorhandling.Exit()
    }()
    // Back to handling the actual input
    if err != nil {
        errorhandling.Log(err, true)
        return
    }
    ui.Log("[ SERVER ] Running on port %d\n", port)
    for {
        var conn net.Conn
        conn, err = ln.Accept()
        activeMu.Lock()
        if !active {
            activeMu.Unlock()
            return
        }
        activeMu.Unlock()
        if err != nil {
            errorhandling.Log(err, false)
            continue
        }
        go handleClient(conn)
    }
}

func handleClient(conn net.Conn) {
    defer conn.Close()
    connAddr := conn.RemoteAddr().String()
    ui.Log("[ %s ] Connected\n", connAddr)
    defer ui.Log("[ %s ] Disconnected\n", connAddr)
    // Read the first request
    sessionID, err := firstRequest(conn)
    if err != nil {
        errorhandling.Log(err, false)
        return
    }
    ui.Log("[ %s ] Attached to Session %x\n", connAddr, sessionID)
    // Now we have to ask for identification.
    ident, err := identRoutine(conn, sessionID)
    if err != nil {
        errorhandling.Log(err, false)
        return
    }
    ui.Log("[ %s ] Identified as `%s`\n", connAddr, ident)
    // Let the manager know who's connected!
    manager.GetManager().AddClient(sessionID, ident, conn)
    // When we leave, let everybody know
    defer func() {
        manager.GetManager().RemoveClient(sessionID, ident, conn)
        leave(sessionID, ident)
    }()
    // Now we have "authenticated" the server.
    // Now the only MTYPEs that actually make sense are CHT(E) and IDENTR
    // We may now begin receiving standard communications
    for {
        msg, err := message.ReadMessage(conn)
        if err == io.EOF {
            // client disconnected
            return
        }
        if err != nil {
            errorhandling.Log(err, false)
            return
        }
        err = handleMessage(conn, msg, sessionID)
        if err != nil {
            errorhandling.Log(err, false)
            return
        }
    }
}

func firstRequest(conn net.Conn) (sessionID uint16, err error) {
    msg, err := message.ReadMessage(conn)
    if err != nil {
        return
    }
    switch msg.MType() {
    case message.JOINR:
        // Parse the message
        var sessionKeyHash []byte
        sessionID, sessionKeyHash = message.ParseJoin(msg.Data())
        // Ask mgr if we can enter
        ok, keyFailed := manager.GetManager().Verify(sessionID, sessionKeyHash) 
        if ok {
            err = sender.SendAcc(conn)
            ui.Log("[ %s ] Accepted to session\n", conn.RemoteAddr().String())
        } else {
            if err = sender.SendReject(conn, keyFailed); err != nil {
                errorhandling.Log(err, false)
            }
            err = fmt.Errorf("invalid login")
        }
        return
    case message.NEWR:
        // Build a new session, if possible
        sessionID, err = manager.GetManager().NewSession(msg.Data())
        if err != nil {
            // that sucks
            return
        } 
        err = sender.SendNew(conn, sessionID)
        ui.Log("[ %s ] Created new session\n", conn.RemoteAddr().String())
        return
    }
    err = fmt.Errorf(
        "Invalid MTYPE for first message: %d",
        msg.MType(),
    )
    errorhandling.Log(err, false)
    return
}

func handleMessage(
    conn net.Conn,
    msg message.Message,
    sessionID uint16,
) (err error) {
    switch msg.MType() {
    case message.IDENTR:
        // Ask the manager for a list of all idents
        var idents [][]byte
        idents, err = manager.GetManager().Identify(sessionID)
        if err != nil {
            return
        }
        // Send the message to the client
        sender.SendIdent(conn, idents)
    case message.CHT, message.CHTE:
        // broadcast the message to the entire session
        var ident []byte
        ident, err = manager.GetManager().GetIdent(sessionID, conn)
        if err != nil {
            errorhandling.Log(err, false)
            return
        }
        alteredMsg := msg.PrependSource(ident)
        manager.GetManager().Broadcast(sessionID, alteredMsg)
    }
    return
}

func identRoutine(conn net.Conn, sessionID uint16) (ident []byte, err error) {
    if err = sender.SendIdentR(conn); err != nil {
        errorhandling.Log(err, false)
        return
    }
    identMsg, err := message.ReadMessage(conn)
    if err != nil {
        errorhandling.Log(err, false)
        return
    }
    idents, err := message.ParseIdent(identMsg.Data())
    if err != nil {
        errorhandling.Log(err, false)
        return
    }
    if len(idents) != 1 {
        err = fmt.Errorf("funny IDENTS: wanted 1, got %d", len(idents))
        errorhandling.Log(err, false)
        return
    }
    ident = idents[0]
    // Tell everybody there's a new friend.
    err = introduce(sessionID, ident)
    if err != nil {
        errorhandling.Log(err, false)
        return
    }
    return
}

func introduce(sessionID uint16, ident []byte) (err error) {
    greeting := fmt.Sprintf("user '%s' has entered the session", ident)
    msg, err := message.NewServerChat([]byte("server"), []byte(greeting))
    if err != nil {
        return err
    }
    manager.GetManager().Broadcast(sessionID, msg)
    return
}

func leave(sessionID uint16, ident []byte) {
    farewell := fmt.Sprintf("user '%s' has exited the session", ident)
    msg, err := message.NewServerChat([]byte("server"), []byte(farewell))
    if err != nil {
        errorhandling.Log(err, false)
        // I would do a sendError, but this is the function for when the user
        // has LEFT, so I can't send anything.
        return
    } 
    manager.GetManager().Broadcast(sessionID, msg)
}
