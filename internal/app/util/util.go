package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"yp-diploma/internal/app/config"
)

func GetRandHexString(lenbyte int) (string, error) {
	newID := make([]byte, lenbyte)
	_, err := rand.Read(newID)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(newID), nil
}

func SignString(msg string) string {
	buf, _ := hex.DecodeString(msg)
	key := sha256.Sum256([]byte(config.PassCiph))
	hm := hmac.New(md5.New, key[:])
	hm.Write(buf)
	sign := hex.EncodeToString(hm.Sum(nil))
	return sign + msg
}

func UnsignString(msg string) (string, bool) {
	buf, err := hex.DecodeString(msg)
	if err != nil {
		return "", false
	}
	key := sha256.Sum256([]byte(config.PassCiph))
	hm := hmac.New(md5.New, key[:])
	sign := buf[:md5.Size]
	hm.Write(buf[md5.Size:])
	newsign := hm.Sum(nil)
	if hmac.Equal(sign, newsign) {
		return msg[2*md5.Size:], true
	}
	return "", false
}

func DecodeString(msg string) (string, error) {
	key := sha256.Sum256([]byte(config.PassCiph))
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	nonce := key[len(key)-aesgcm.NonceSize():]

	buf, err := hex.DecodeString(msg)
	if err != nil {
		return "", err
	}

	encbuf, err := aesgcm.Open(nil, nonce, buf, nil)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(encbuf), nil
}

func EncodeString(msg string) (string, error) {
	key := sha256.Sum256([]byte(config.PassCiph))
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	nonce := key[len(key)-aesgcm.NonceSize():]

	buf, err := hex.DecodeString(msg)
	if err != nil {
		return "", err
	}

	encbuf := aesgcm.Seal(nil, nonce, buf, nil)

	return hex.EncodeToString(encbuf), nil
}
