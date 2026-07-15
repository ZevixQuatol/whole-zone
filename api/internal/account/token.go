package account

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

const tokenBytes = 32

type Token struct {
	Raw  string
	Hash []byte
}

func NewToken() (Token, error) {
	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return Token{}, err
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	return Token{Raw: encoded, Hash: TokenHash(encoded)}, nil
}

func ParseToken(raw string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(decoded) != tokenBytes {
		return nil, errors.New("invalid token format")
	}
	return TokenHash(raw), nil
}

func TokenHash(raw string) []byte {
	hash := sha256.Sum256([]byte(raw))
	return hash[:]
}
