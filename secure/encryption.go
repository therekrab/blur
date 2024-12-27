package secure

import (
    "crypto/rand"
    "crypto/cipher"
    "crypto/aes"
)

const KEYSIZE int = 32

func GenKey(key string) (validKey []byte) {
    if len(key) == 0 {
        return make([]byte, KEYSIZE, KEYSIZE)
    }
    if len(key) >= KEYSIZE {
        validKey = []byte(key[:KEYSIZE])
        return
    }
    return GenKey(key + key)
}

func BuildAesGCM(key []byte) (aesGCM cipher.AEAD, err error){
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    aesGCM, err = cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    return
}


func EncryptData(aesGCM cipher.AEAD, data []byte) (out []byte, err error) {
    // Generate a nonce
    nonce := make([]byte, 12)
    if _, err := rand.Read(nonce); err != nil {
        return nil, err
    }
    // actually do some encrypting.
    encrypted := aesGCM.Seal(nil, nonce, data, nil)
    out = append(nonce, encrypted...)
    return
}

func DecryptData(aesGCM cipher.AEAD, data []byte) (out []byte, err error) {
    nonce := data[:12]
    encryptedData := data[12:]
    // now decrypt it!
    return aesGCM.Open(nil, nonce, encryptedData, nil)
}
