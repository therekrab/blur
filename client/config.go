package client

import (
	"crypto/cipher"
	"github.com/therekrab/blur/secure"
)

type ClientConfig struct {
    sessionID uint16
    ident []byte
    key []byte
    join bool
    aesGCM cipher.AEAD
}

func (cc *ClientConfig) HashedKey() []byte {
    hk := secure.Hash(cc.key)
    return hk
}

func (cc *ClientConfig) encrypt(msg string) (encrypted []byte, err error) {
    return secure.EncryptData(cc.aesGCM, []byte(msg))
}

func (cc *ClientConfig) decrypt(encrypted []byte) (msg []byte, err error) {
    return secure.DecryptData(cc.aesGCM, encrypted)
}

func JoinSessionConfig(
    sessionID uint16,
    sessionKey string,
    ident string,
) (client ClientConfig, err error) {
    newKey := secure.GenKey(sessionKey)
    var aesGCM cipher.AEAD
    aesGCM, err = secure.BuildAesGCM(newKey)
    if err != nil {
        return
    }
    client = ClientConfig {
        sessionID,
        []byte(ident),
        newKey,
        true,
        aesGCM,
    }
    return
}

func NewSessionConfig(
    sessionKey string,
    ident string,
) (client ClientConfig, err error) {
    newKey := secure.GenKey(sessionKey)
    var aesGCM cipher.AEAD
    aesGCM, err = secure.BuildAesGCM(newKey)
    if err != nil {
        return
    }
    client = ClientConfig {
        0, // This will be set later.
        []byte(ident),
        newKey,
        false,
        aesGCM,
    }
    return
}
