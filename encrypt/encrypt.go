package encrypt

import (
	"crypto/cipher"
)

type DecOrEnc int

const (
	Decrypt DecOrEnc = iota
	Encrypt
)

type cipherInfo struct {
	keyLen    int
	ivLen     int
	newStream func(key, iv []byte, doe DecOrEnc) (cipher.Stream, error)
}

type Cipher struct {
	enc  cipher.Stream
	dec  cipher.Stream
	key  []byte
	info *cipherInfo
	ota  bool // one-time auth
	iv   []byte
}
