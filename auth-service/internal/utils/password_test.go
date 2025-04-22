package utils

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	const pwd = "P@ssw0rd!"

	hash, err := HashPassword(pwd)
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}
	if hash == pwd {
		t.Fatalf("hash should differ from plain text")
	}

	if ok := CheckPasswordHash(pwd, hash); !ok {
		t.Fatalf("hash mismatch, expected true")
	}
}
