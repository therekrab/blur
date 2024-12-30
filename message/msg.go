package message

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
)

type Message struct {
    dsize uint16
    mtype MType
    data []byte
}

func (msg *Message) DSize() uint16 {
    return msg.dsize
}

func (msg *Message) MType() MType {
    return msg.mtype
}

func (msg *Message) Data() []byte {
    return msg.data
}

func NewMessage(dsize uint16, mtype MType, data []byte) (msg Message) {
    return Message{dsize, mtype, data} 
}

func (msg *Message) bytes() []byte {
    msgBytes := make([]byte, 2)
    binary.BigEndian.PutUint16(msgBytes, msg.dsize)
    msgBytes = append(msgBytes, byte(msg.mtype))
    msgBytes = append(msgBytes, msg.data...)
    return msgBytes
}

func (msg *Message) SendTo(conn net.Conn) (err error) {
    msgBytes := msg.bytes()
    _, err = conn.Write(msgBytes)
    if err != nil {
        return err
    }
    return
}

func NewChat(data []byte) (msg Message, err error) {
    dsize := len(data)
    if dsize > math.MaxUint16 {
        err = fmt.Errorf("message was too large")
    }
    msg = NewMessage(uint16(dsize), CHT, data)
    return
}

func NewServerChat(source []byte, data []byte) (msg Message, err error) {
    msg, err = NewChat(data)
    if err != nil {
        return
    }
    msg = msg.PrependSource(source)
    return
}

func NewIdent(idents [][]byte) (msg Message, err error) {
    data := make([]byte, 0)
    for _, ident := range idents {
        dsize := uint16(len(ident))
        dsizeBytes := make([]byte, 2)
        binary.BigEndian.PutUint16(dsizeBytes, dsize)
        data = append(data, dsizeBytes...)
        data = append(data, ident...)
    }
    terminator := make([]byte, 2, 2)
    data = append(data, terminator...)
    msg = NewMessage(uint16(len(data)), IDENT, data)
    return
}

func ReadMessage(conn net.Conn) (msg Message, err error) {
    dsizeBytes := make([]byte, 2)
    n, err := conn.Read(dsizeBytes)
    if err != nil {
        return
    }
    if n != 2 {
        err = fmt.Errorf("could not read DSIZE")
        return
    }
    dsize := binary.BigEndian.Uint16(dsizeBytes)
    mtypeBytes := make([]byte, 1)
    n, err = conn.Read(mtypeBytes)
    if err != nil {
        return
    }
    if n != 1 {
        err = fmt.Errorf("could not read MTYPE")
        return
    }
    mtype := MType(mtypeBytes[0])
    data := make([]byte, dsize)
    n, err = conn.Read(data)
    if err != nil {
        return
    }
    if uint16(n) != dsize {
        err = fmt.Errorf(
            "could not read DATA: expected %d bytes, got %d.",
            dsize,
            n,
        )
    }
    msg = NewMessage(dsize, mtype, data)
    return

}

func (msg *Message) PrependSource(ident []byte) (alteredMsg Message) {
    miniDsize := uint16(len(ident))
    alteredDsize := msg.dsize + miniDsize + 2 // 2 for the size of miniDsize.
    miniDsizeBytes := make([]byte, 2)
    binary.BigEndian.PutUint16(miniDsizeBytes, miniDsize)
    miniData := append(miniDsizeBytes, ident...)
    alteredData := append(miniData, msg.data...)
    alteredMsg = NewMessage(alteredDsize, msg.mtype, alteredData)
    return
}

