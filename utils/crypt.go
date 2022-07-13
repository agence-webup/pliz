// From https://github.com/Xeoncross/go-aesctr-with-hmac
// Author Xeoncross
// File https://github.com/Xeoncross/go-aesctr-with-hmac/blob/master/crypt.go
// Commit a777569d9869525dbd110ad743b2b658dfc701c5

package utils

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/scrypt"
)

// hmacSize must be less to BUFFER_SIZE
const BUFFER_SIZE int = 16 * 1024
const IV_SIZE int = 16
const SALT_SIZE int = 32
const V1 byte = 0x1
const hmacSize = sha512.Size

// ErrInvalidHMAC for authentication failure
var ErrInvalidHMAC = errors.New("Invalid HMAC")

// Encrypt the stream using the given AES-CTR and SHA512-HMAC key
func Encrypt(in io.Reader, out io.Writer, key []byte) (err error) {
	keyAes, saltAes, err := DeriveKey(key, nil)
	if err != nil {
		return err
	}
	keyHmac, saltHmac, err := DeriveKey(key, nil)
	if err != nil {
		return err
	}

	iv := make([]byte, IV_SIZE)
	_, err = rand.Read(iv)
	if err != nil {
		return err
	}

	AES, err := aes.NewCipher(keyAes)
	if err != nil {
		return err
	}

	ctr := cipher.NewCTR(AES, iv)
	HMAC := hmac.New(sha512.New, keyHmac) // https://golang.org/pkg/crypto/hmac/#New

	// Version
	_, err = out.Write([]byte{V1})
	if err != nil {
		return err
	}

	w := io.MultiWriter(out, HMAC)

	_, err = w.Write(iv)
	if err != nil {
		return err
	}
	_, err = w.Write(saltAes)
	if err != nil {
		return err
	}
	_, err = w.Write(saltHmac)
	if err != nil {
		return err
	}

	buf := make([]byte, BUFFER_SIZE)
	for {
		n, err := in.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		if n != 0 {
			outBuf := make([]byte, n)
			ctr.XORKeyStream(outBuf, buf[:n])
			_, err = w.Write(outBuf)
			if err != nil {
				return err
			}
		}

		if err == io.EOF {
			break
		}
	}

	_, err = out.Write(HMAC.Sum(nil))

	return err
}

// Decrypt the stream and verify HMAC using the given AES-CTR and SHA512-HMAC key
// Do not trust the out io.Writer contents until the function returns the result
// of validating the ending HMAC hash.
func Decrypt(in io.Reader, out io.Writer, key []byte) (err error) {

	// Read version (up to 0-255)
	var version int8
	err = binary.Read(in, binary.LittleEndian, &version)
	if err != nil {
		return err
	}

	iv := make([]byte, IV_SIZE)
	_, err = io.ReadFull(in, iv)
	if err != nil {
		return err
	}
	saltAes := make([]byte, SALT_SIZE)
	_, err = io.ReadFull(in, saltAes)
	if err != nil {
		return err
	}
	saltHmac := make([]byte, SALT_SIZE)
	_, err = io.ReadFull(in, saltHmac)
	if err != nil {
		return err
	}
	keyAes, _, err := DeriveKey(key, saltAes)
	if err != nil {
		return err
	}
	keyHmac, _, err := DeriveKey(key, saltHmac)
	if err != nil {
		return err
	}

	AES, err := aes.NewCipher(keyAes)
	if err != nil {
		return err
	}

	ctr := cipher.NewCTR(AES, iv)
	h := hmac.New(sha512.New, keyHmac)
	h.Write(iv)
	h.Write(saltAes)
	h.Write(saltHmac)
	mac := make([]byte, hmacSize)

	w := out

	buf := bufio.NewReaderSize(in, BUFFER_SIZE)
	var limit int
	var b []byte
	for {
		b, err = buf.Peek(BUFFER_SIZE)
		if err != nil && err != io.EOF {
			return err
		}

		limit = len(b) - hmacSize

		// We reached the end
		if err == io.EOF {

			left := buf.Buffered()
			if left < hmacSize {
				return errors.New("not enough left")
			}

			copy(mac, b[left-hmacSize:left])

			if left == hmacSize {
				break
			}
		}

		h.Write(b[:limit])

		// We always leave at least hmacSize bytes left in the buffer
		// That way, our next Peek() might be EOF, but we will still have enough
		outBuf := make([]byte, int64(limit))
		_, err = buf.Read(b[:limit])
		if err != nil {
			return err
		}
		ctr.XORKeyStream(outBuf, b[:limit])
		_, err = w.Write(outBuf)
		if err != nil {
			return err
		}

		if err == io.EOF {
			break
		}
	}

	if !hmac.Equal(mac, h.Sum(nil)) {
		return ErrInvalidHMAC
	}

	return nil
}

func DeriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, SALT_SIZE)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}

	key, err := scrypt.Key(password, salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, nil, err
	}

	return key, salt, nil
}
