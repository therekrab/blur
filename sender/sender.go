package sender

import (
	"encoding/binary"
	"fmt"
	"net"
	"therekrab/secrets/errorhandling"
	"therekrab/secrets/message"
)

// The only reason that the error is handled inside of send is because the
// caller of sendError() already HAS an error - the reason for calling the
// function. So we will never OVERRIDE errors, so we just handle it in here.
func SendError(conn net.Conn) {
    fmt.Println("sending ERR")
    errorMsg := message.NewMessage(0, message.ERR, nil)
    err := errorMsg.SendTo(conn)
    if err != nil {
        errorhandling.Report(err, false)
    }
}

func SendReject(conn net.Conn, keyFailed bool) (err error) {
    fmt.Println("sending reject")
    reason := make([]byte, 1)
    if keyFailed {
        reason[0] = 1
    }
    rejMsg := message.NewMessage(1, message.REJ, reason)
    err = rejMsg.SendTo(conn)
    return
}

func SendAcc(conn net.Conn) (err error) {
    accMsg := message.NewMessage(0, message.ACC, nil)
    err = accMsg.SendTo(conn)
    return
}

func SendNew(conn net.Conn, sessionID uint16) (err error) {
    sessionIDBytes := make([]byte, 2)
    binary.BigEndian.PutUint16(sessionIDBytes, sessionID)
    newMsg := message.NewMessage(2, message.NEW, sessionIDBytes)
    err = newMsg.SendTo(conn)
    return
}

func SendNewR(conn net.Conn, sessionKeyHashed []byte) (err error) {
    fmt.Printf("len(hashed key) = %d\n", len(sessionKeyHashed))
    newRMsg := message.NewMessage(
        uint16(len(sessionKeyHashed)),
        message.NEWR,
        sessionKeyHashed,
    )
    err = newRMsg.SendTo(conn)
    return
}

func SendIdentR(conn net.Conn) (err error) {
    identRMsg := message.NewMessage(0, message.IDENTR, nil)
    err = identRMsg.SendTo(conn)
    return
}

func SendIdent(conn net.Conn, idents [][]byte) (err error) {
    identMsg, err := message.NewIdent(idents)
    if err != nil {
        return err
    }
    err = identMsg.SendTo(conn)
    return
}

func SendJoinR(
    conn net.Conn,
    sessionID uint16,
    sessionKeyHash []byte,
) (err error) {
    data := make([]byte, 2)
    binary.BigEndian.PutUint16(data, sessionID)
    data = append(data, sessionKeyHash...)
    joinMsg := message.NewMessage(
        uint16(len(sessionKeyHash) + 2),
        message.JOINR,
        data,
    )
    err = joinMsg.SendTo(conn)
    return
}

func SendChatE(conn net.Conn, data []byte) (err error) {
    chtEMsg := message.NewMessage(
        uint16(len(data)),
        message.CHTE,
        data,
    )
    err = chtEMsg.SendTo(conn)
    return
}
