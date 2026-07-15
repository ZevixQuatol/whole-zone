package account

import (
	"bytes"
	"testing"
)

func TestTokenRoundTrip(t *testing.T) {
	first, err := NewToken()
	if err != nil {
		t.Fatalf("new token: %v", err)
	}
	second, err := NewToken()
	if err != nil {
		t.Fatalf("new second token: %v", err)
	}
	if first.Raw == second.Raw {
		t.Fatal("generated duplicate tokens")
	}
	if first.Raw == string(first.Hash) {
		t.Fatal("raw token and persisted hash are identical")
	}

	hash, err := ParseToken(first.Raw)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if !bytes.Equal(hash, first.Hash) {
		t.Fatal("parsed token hash does not match")
	}
}

func TestTokenRejectsInvalidEncodingAndLength(t *testing.T) {
	for _, raw := range []string{"", "not base64!", "c2hvcnQ"} {
		if _, err := ParseToken(raw); err == nil {
			t.Fatalf("expected %q to fail", raw)
		}
	}
}
