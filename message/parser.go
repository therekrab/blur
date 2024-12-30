package message

import (
	"encoding/binary"
	"fmt"
)

func ParseJoin(data []byte) (sessionID uint16, sessionKeyHash []byte) {
    sessionIDBytes := data[:2] // First two bytes
    sessionKeyHash = data[2:]
    sessionID = binary.BigEndian.Uint16(sessionIDBytes)
    return
}

func ParseIdent(data []byte) (idents [][]byte, err error) {
    idents = make([][]byte, 0)
    i := 0
    for i < len(data) {
        // Read two bytes!
        if i + 2 > len(data) {
            err = fmt.Errorf("invalid IDENT response")
            return
        }
        dSizeBytes := data[i:i+2]
        i += 2

        dSize := int(binary.BigEndian.Uint16(dSizeBytes))
        if dSize == 0 {
            break
        }

        if i + dSize >= len(data) {
            err = fmt.Errorf("invalid IDENT response")
            return
        }
        
        identBytes := data[i:i+dSize]
        idents = append(idents, identBytes)
        i += dSize
    }
    return 
}

func ParseCht(data []byte) (source []byte, cht []byte, err error) {
    if len(data) < 2 {
        err = fmt.Errorf("CHT DATA too short")
        return
    }
    dsizeBytes := data[:2]
    dsize := binary.BigEndian.Uint16(dsizeBytes)
    if len(data) < 2 + int(dsize) {
        err = fmt.Errorf("CHT DATA too short")
        return
    }
    source = data[2:2+dsize]
    cht = data[2+dsize:]
    return
}
