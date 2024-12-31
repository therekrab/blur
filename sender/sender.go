package sender

import (
	"encoding/binary"
	"net"
	"therekrab/secrets/message"
)

func SendReject(conn net.Conn, keyFailed bool) (err error) {
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
