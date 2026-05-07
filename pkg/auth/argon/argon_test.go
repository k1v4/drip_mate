package argon

import (
	"strings"
	"testing"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	if params.Time != 1 {
		t.Errorf("expected Time = 1, got %d", params.Time)
	}

	if params.Memory != 64*1024 {
		t.Errorf("expected Memory = 65536, got %d", params.Memory)
	}

	if params.Threads != 4 {
		t.Errorf("expected Threads = 4, got %d", params.Threads)
	}

	if params.KeyLen != 32 {
		t.Errorf("expected KeyLen = 32, got %d", params.KeyLen)
	}

	if params.SaltLen != 16 {
		t.Errorf("expected SaltLen = 16, got %d", params.SaltLen)
	}
}

func TestNewArgon2Hasher(t *testing.T) {
	params := DefaultParams()

	hasher := NewArgon2Hasher(params, "pepper")

	if hasher == nil {
		t.Fatal("expected hasher, got nil")
	}

	if hasher.params != params {
		t.Error("params were not assigned")
	}

	if hasher.pepper != "pepper" {
		t.Errorf("expected pepper 'pepper', got '%s'", hasher.pepper)
	}
}

func TestHasher_Hash(t *testing.T) {
	hasher := NewArgon2Hasher(DefaultParams(), "pepper")

	hash, err := hasher.Hash("password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("unexpected hash format: %s", hash)
	}
}

func TestHasher_Hash_DifferentSalts(t *testing.T) {
	hasher := NewArgon2Hasher(DefaultParams(), "pepper")

	hash1, err := hasher.Hash("password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hash2, err := hasher.Hash("password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected hashes to differ because of random salt")
	}
}

func TestHasher_Verify_Success(t *testing.T) {
	hasher := NewArgon2Hasher(DefaultParams(), "pepper")

	password := "password123"

	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ok, err := hasher.Verify(password, hash)
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}

	if !ok {
		t.Fatal("expected password verification to succeed")
	}
}

func TestHasher_Verify_WrongPassword(t *testing.T) {
	hasher := NewArgon2Hasher(DefaultParams(), "pepper")

	hash, err := hasher.Hash("correct-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ok, err := hasher.Verify("wrong-password", hash)
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}

	if ok {
		t.Fatal("expected verification to fail")
	}
}

func TestHasher_Verify_InvalidFormat(t *testing.T) {
	hasher := NewArgon2Hasher(DefaultParams(), "pepper")

	ok, err := hasher.Verify("password", "invalid-hash")

	if err == nil {
		t.Fatal("expected format error")
	}

	if ok {
		t.Fatal("expected verification failure")
	}
}

func TestHasher_Verify_InvalidParams(t *testing.T) {
	hasher := NewArgon2Hasher(DefaultParams(), "pepper")

	hash := "$argon2id$v=19$invalid$params$hash"

	ok, err := hasher.Verify("password", hash)

	if err == nil {
		t.Fatal("expected params parsing error")
	}

	if ok {
		t.Fatal("expected verification failure")
	}
}

func TestHasher_Verify_InvalidSalt(t *testing.T) {
	hasher := NewArgon2Hasher(DefaultParams(), "pepper")

	hash := "$argon2id$v=19$m=65536,t=1,p=4$invalid-base64$$"

	ok, err := hasher.Verify("password", hash)

	if err == nil {
		t.Fatal("expected invalid salt error")
	}

	if ok {
		t.Fatal("expected verification failure")
	}
}

func TestHasher_Verify_InvalidHash(t *testing.T) {
	hasher := NewArgon2Hasher(DefaultParams(), "pepper")

	hash := "$argon2id$v=19$m=65536,t=1,p=4$c2FsdA$invalid-base64"

	ok, err := hasher.Verify("password", hash)

	if err == nil {
		t.Fatal("expected invalid hash error")
	}

	if ok {
		t.Fatal("expected verification failure")
	}
}

func TestHasher_Verify_WrongPepper(t *testing.T) {
	params := DefaultParams()

	hasher1 := NewArgon2Hasher(params, "pepper-1")
	hasher2 := NewArgon2Hasher(params, "pepper-2")

	password := "password123"

	hash, err := hasher1.Hash(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ok, err := hasher2.Verify(password, hash)
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}

	if ok {
		t.Fatal("expected verification to fail with different pepper")
	}
}
