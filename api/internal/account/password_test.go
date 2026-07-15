package account

import (
	"strings"
	"testing"
)

func TestPasswordRoundTrip(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if strings.Contains(hash, "correct horse") {
		t.Fatal("password hash contains plaintext")
	}

	ok, err := CheckPassword("correct horse battery staple", hash)
	if err != nil || !ok {
		t.Fatalf("check correct password: ok=%v err=%v", ok, err)
	}
	ok, err = CheckPassword("wrong password", hash)
	if err != nil || ok {
		t.Fatalf("check wrong password: ok=%v err=%v", ok, err)
	}
}

func TestPasswordRejectsMalformedHash(t *testing.T) {
	if _, err := CheckPassword("password", "not-an-argon2-hash"); err == nil {
		t.Fatal("expected malformed hash to fail")
	}
}
