package secure

import (
	"crypto/sha256"
)

func Hash(sessionKey []byte) []byte {
    h := sha256.New()
    h.Write(sessionKey)
    return h.Sum(nil)
}
