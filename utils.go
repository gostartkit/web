package web

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

var (
	htmlQuot = []byte("&#34;") // shorter than "&quot;"
	htmlApos = []byte("&#39;") // shorter than "&apos;" and apos was not in HTML until HTML5
	htmlAmp  = []byte("&amp;")
	htmlLt   = []byte("&lt;")
	htmlGt   = []byte("&gt;")
)

const (
	ENV_VIEW_DIR         = "AFXCN_WEB_VIEW_DIR"
	ENV_COOKIE_SECRET    = "AFXCN_WEB_COOKIE_SECRET"
	ENV_COOKIE_ENC_SALT  = "AFXCN_WEB_COOKIE_ENC_SALT"
	ENV_COOKIE_SIGN_SALT = "AFXCN_WEB_COOKIE_SIGN_SALT"
	ENV_DRIVER_NAME      = "AFXCN_WEB_DRIVER_NAME"
	ENV_DATA_SOURCE_NAME = "AFXCN_WEB_DATA_SOURCE_NAME"
)

func Getenv(key string) string {
	return os.Getenv(key)
}

func Setenv(key string, value string) error {
	return os.Setenv(key, value)
}

func htmlEscape(w io.Writer, b []byte) {
	last := 0
	for i, c := range b {
		var html []byte
		switch c {
		case '"':
			html = htmlQuot
		case '\'':
			html = htmlApos
		case '&':
			html = htmlAmp
		case '<':
			html = htmlLt
		case '>':
			html = htmlGt
		default:
			continue
		}
		w.Write(b[last:i])
		w.Write(html)
		last = i + 1
	}
	w.Write(b[last:])
}

func contentType(val string) string {
	var ctype string
	if strings.ContainsRune(val, '/') {
		ctype = val
	} else {
		if !strings.HasPrefix(val, ".") {
			val = "." + val
		}
		ctype = mime.TypeByExtension(val)
	}
	return ctype
}

func newCookie(name string, value string, age int64) *http.Cookie {
	var utctime time.Time
	if age == 0 {
		utctime = time.Unix(2147483647, 0)
	} else {
		utctime = time.Unix(time.Now().Unix()+age, 0)
	}
	return &http.Cookie{Name: name, Value: value, Expires: utctime}
}

func genKey(password string, salt string) []byte {
	return pbkdf2.Key([]byte(password), []byte(salt), pbkdf2Iterations, keySize, sha256.New)
}

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesCipher, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}

func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	if len(ciphertext) <= aes.BlockSize {
		return nil, errors.New("Invalid cipher text")
	}
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plaintext := make([]byte, len(ciphertext)-aes.BlockSize)
	stream := cipher.NewCTR(aesCipher, ciphertext[:aes.BlockSize])
	stream.XORKeyStream(plaintext, ciphertext[aes.BlockSize:])
	return plaintext, nil
}

func sign(data []byte, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func randString(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)

	if err != nil {
		return ""
	}

	return string(b)
}

func envOrRandom(key string, n int) string {
	val := Getenv(key)

	if len(val) == 0 {
		val = randString(n)
	}

	return val
}
