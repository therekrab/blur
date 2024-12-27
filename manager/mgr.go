package manager

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"sync"
	"therekrab/secrets/errorhandling"
	"therekrab/secrets/message"
)

type Manager struct {
    smgrs map[uint16]sessionManager
    mu sync.Mutex
}

var mgr *Manager
var once sync.Once

func GetManager() *Manager {
    once.Do(func() {
        mgr = &Manager{}
        mgr.smgrs = make(map[uint16]sessionManager, 0)
    })
    return mgr
}

func (mgr *Manager) getSessionManager(sessionID uint16) *sessionManager {
    smgr, ok := mgr.smgrs[sessionID]
    if !ok {
        return nil
    }
    return &smgr
}

func (mgr *Manager) AddClient(sessionID uint16, ident []byte, conn net.Conn) {
    mgr.mu.Lock() // gotta be safe
    defer mgr.mu.Unlock() // gotta avoid dead locks
    smgr := mgr.getSessionManager(sessionID)
    if smgr == nil {
        err := fmt.Errorf("invalid session ID: %x", sessionID)
        errorhandling.Report(err, false)
        return
    }
    smgr.addClient(conn, ident)
    fmt.Printf("New client '%s' on session %x\n", ident, sessionID)
}

func (mgr *Manager) RemoveClient(sessionID uint16, conn net.Conn) {
    mgr.mu.Lock()
    defer mgr.mu.Unlock()
    smgr := mgr.getSessionManager(sessionID)
    if smgr == nil {
        err := fmt.Errorf("invalid session ID: %x", sessionID)
        errorhandling.Report(err, false)
        return
    }
    smgr.removeClient(conn)
    fmt.Printf("Client left session %x\n", sessionID)
}

func (mgr *Manager) Broadcast(
    sessionID uint16,
    msg message.Message,
    exclude []byte,
) (err error) {
    mgr.mu.Lock()
    defer mgr.mu.Unlock()
    smgr := mgr.getSessionManager(sessionID)
    if smgr == nil {
        err = fmt.Errorf("invalid sessionID for broadcast")
        return
    }
    smgr.broadcast(msg, exclude)
    return
}

func (mgr *Manager) newSessionID() (sessionID uint16, err error) {
    if len(mgr.smgrs) > math.MaxUint16 / 3 * 2 {
        // If we're over two-thirds full, we won't take any more clients
        // This prevents us from filling up the server, and taking forever
        // to assign a new session ID.
        err = fmt.Errorf("too many connections")
        return
    }
    for {
        sessionIDBytes := make([]byte, 2) // 2 bytes = 16 bits
        _, err = rand.Read(sessionIDBytes)
        if err != nil {
            return
        }
        sessionID = binary.BigEndian.Uint16(sessionIDBytes)
        if _, ok := mgr.smgrs[sessionID]; !ok {
            // we can leave this loop!
            return
        }
    }
}

func (mgr *Manager) NewSession(sessionKeyHash []byte) (sessionID uint16, err error) {
    mgr.mu.Lock()
    defer mgr.mu.Unlock()
    sessionID, err = mgr.newSessionID()
    if err != nil {
        return
    }
    mgr.smgrs[sessionID] = newSessionManager(sessionKeyHash)
    return
}

func (mgr *Manager) Verify(
    sessionID uint16,
    sessionKeyHash []byte,
) (ok bool, keyFailed bool) {
    mgr.mu.Lock()
    defer mgr.mu.Unlock()
    if smgr, found := mgr.smgrs[sessionID]; found {
        ok = smgr.verify(sessionKeyHash)
        return ok, !ok
    }
    fmt.Printf("bad session ID: %x\n", sessionID)
    return
}

func (mgr *Manager) Identify(sessionID uint16) (idents [][]byte, err error) {
    mgr.mu.Lock()
    defer mgr.mu.Unlock()
    if smgr, ok := mgr.smgrs[sessionID]; ok {
        idents = smgr.identify()
        return
    }
    err = fmt.Errorf("invalid sessionID for identification request")
    return
}

func (mgr *Manager) GetIdent(sessionID uint16, conn net.Conn) (ident []byte, err error) {
    mgr.mu.Lock()
    defer mgr.mu.Unlock()
    if smgr, ok := mgr.smgrs[sessionID]; ok {
        ident, err = smgr.getIdent(conn)
        return
    }
    err = fmt.Errorf("invalid session ID for getIdent request")
    return
}
