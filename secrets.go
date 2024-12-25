package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func loadBytes(file io.Reader) (bytes []byte) {
    _, err := file.Read(bytes)
    if err != nil {
        fmt.Fprintf(os.Stderr, "error reading from file: %s\n", err)
        os.Exit(1)
    }
    return
}

func buildAesGCM(key []byte) (aesGCM cipher.AEAD, err error){
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

func genKey(key string) (validKey []byte) {
    if len(key) >= 32 {
        validKey = []byte(key[:32])
        return
    }
    return genKey(key + key)
}

func encryptFile(aesGCM cipher.AEAD, filepath string) (err error) {
    // open the input file for reading
    file, err := os.Open(filepath)
    if err != nil {
        return err
    }
    defer file.Close()
    // load the bytes that we need to encrypt.
    data, err := io.ReadAll(file)
    if err != nil {
        return err
    }
    // Generate a nonce
    nonce := make([]byte, 12)
    if _, err := rand.Read(nonce); err != nil {
        return err
    }
    // actually do some encrypting.
    encrypted := aesGCM.Seal(nil, nonce, data, nil)
    // open the output file
    outputFilepath := filepath + ".secret"
    outputFile, err := os.Create(outputFilepath)
    if err != nil {
        return err 
    }
    defer outputFile.Close()
    // save the nonce and encrypted data to the output file.
    // For this implementation, we simply write the nonce,
    // then append all the encrypted bytes.
    allData := append(nonce, encrypted...)
    if _, err = outputFile.Write(allData); err != nil {
        return err
    }
    return
}

func decryptFile(aesGCM cipher.AEAD, filepath string) (err error) {
    // open the input data
    file, err := os.Open(filepath)
    if err != nil {
        return err
    }
    defer file.Close()
    // load the encrypted data (and the nonce)
    nonce := make([]byte, 12)
    if _, err := file.Read(nonce); err != nil {
        return err
    }
    encrypedData, err := io.ReadAll(file)
    if err != nil {
        return err
    }
    // now decrypt it!
    data, err := aesGCM.Open(nil, nonce, encrypedData, nil)
    if err != nil {
        return err
    }
    // write data to output file
    outputFilepath := strings.TrimSuffix(filepath, ".secret")
    outputFile, err := os.Create(outputFilepath)
    if err != nil {
        return err
    }
    defer outputFile.Close()
    if _, err = outputFile.Write(data); err != nil {
        return err
    }
    return
}

func main() {
    var decryptFlag bool
    flag.BoolVar(&decryptFlag, "u", false, "enables decryption")
    flag.Parse()

    filepaths := flag.Args()
    if len(filepaths) == 0 {
        fmt.Fprintln(os.Stderr, "no filepaths supplied!")
        os.Exit(2)
    }
    // get the password
    fmt.Print("Enter password: ")
    
    rdr := bufio.NewReader(os.Stdin)
    passwd, err := rdr.ReadString('\n')
    if err != nil {
        fmt.Fprintf(os.Stderr, "error reading from stdin: %s\n", err)
        os.Exit(1)
    }
    key := genKey(strings.TrimSpace(passwd))
    // Build the AES-GCM cipher
    aesGCM, err := buildAesGCM(key)
    if err != nil {
        fmt.Fprintf(os.Stderr, "error building AES-GCM: %s\n", err)
        os.Exit(1)
    }
    if decryptFlag {
        // encryption mode
        for _, filepath := range filepaths {
            if err = decryptFile(aesGCM, filepath); err != nil {
                fmt.Fprintf(
                    os.Stderr,
                    "error decrypting %s: %s\n",
                    filepath,
                    err)
                } else {
                    fmt.Printf("successfully decrypted %s\n", filepath)
                }
            }
    } else {
        // decryption mode
        for _, filepath := range filepaths {
            if err = encryptFile(aesGCM, filepath); err != nil {
                fmt.Fprintf(
                    os.Stderr,
                    "error encrypting %s: %s\n",
                    filepath,
                    err)
            } else {
                fmt.Printf("successfully encrypted %s\n", filepath)
            }
        }
    }
}
