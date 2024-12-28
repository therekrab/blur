package manager

import (
	"fmt"
	"net"
	"slices"
	"therekrab/secrets/errorhandling"
	"therekrab/secrets/message"
)

type sessionManager struct {
    clients map[net.Conn][]byte
    sessionKeyHash []byte
}

func newSessionManager(sessionKeyHash []byte) (smgr sessionManager) {
    smgr.clients = make(map[net.Conn][]byte)
    smgr.sessionKeyHash = sessionKeyHash
    return
}

func (smgr *sessionManager) addClient(conn net.Conn, ident []byte) {
    smgr.clients[conn] = ident
}

func (smgr *sessionManager) removeClient(conn net.Conn) {
    delete(smgr.clients, conn)
}

func (smgr *sessionManager) isEmpty() bool {
    return len(smgr.clients) == 0
}

func (smgr *sessionManager) broadcast(msg message.Message, exclude []byte) {
    for conn, ident := range smgr.clients {
        if slices.Equal(ident, exclude) {
            // this is probably the sender.
            continue
        }
        err := msg.SendTo(conn)
        if err != nil {
            addr := conn.LocalAddr().String()
            err := fmt.Errorf("could not send data to %s: %s", addr, err)
            errorhandling.Report(err, false)
        }
    }
}

func (smgr *sessionManager) verify(sessionKeyHash []byte) bool {
    return slices.Equal(smgr.sessionKeyHash, sessionKeyHash)
}

func (smgr *sessionManager) identify() (idents [][]byte) {
    idents = make([][]byte, 0)
    for _, ident := range smgr.clients {
        idents = append(idents, ident)
    }
    return
}

func (smgr *sessionManager) getIdent(conn net.Conn) (ident []byte, err error) {
    if ident, ok := smgr.clients[conn]; ok {
        return ident, nil
    }
    err = fmt.Errorf("conn not in client list")
    return
}
