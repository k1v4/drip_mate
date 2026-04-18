package argon

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Params struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen uint32
}

type Hasher struct {
	params *Params
	pepper string
}

func DefaultParams() *Params {
	return &Params{
		Time:    1,
		Memory:  64 * 1024, // 64MB
		Threads: 4,
		KeyLen:  32,
		SaltLen: 16,
	}
}

func NewArgon2Hasher(params *Params, pepper string) *Hasher {
	return &Hasher{
		params: params,
		pepper: pepper,
	}
}

func (a *Hasher) Hash(password string) (string, error) {
	salt := make([]byte, a.params.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// 👇 добавляем pepper
	input := password + a.pepper

	hash := argon2.IDKey(
		[]byte(input),
		salt,
		a.params.Time,
		a.params.Memory,
		a.params.Threads,
		a.params.KeyLen,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		a.params.Memory,
		a.params.Time,
		a.params.Threads,
		b64Salt,
		b64Hash,
	)

	return encoded, nil
}

func (a *Hasher) Verify(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format: expected 6 parts, got %d", len(parts))
	}

	var memory, time uint32
	var threads uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, fmt.Errorf("invalid params segment %q: %w", parts[3], err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("invalid salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("invalid hash: %w", err)
	}

	input := password + a.pepper

	hash := argon2.IDKey(
		[]byte(input),
		salt,
		time,
		memory,
		threads,
		uint32(len(expectedHash)),
	)

	if subtle.ConstantTimeCompare(hash, expectedHash) == 1 {
		return true, nil
	}

	return false, nil
}
