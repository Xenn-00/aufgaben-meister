package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

func GenerateHash(payload string) (string, error) {
	// Hier sollte die eigentliche Hashing-Logik implementiert werden.
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	t := uint32(3)
	memory := uint32(64 * 1024)
	threads := uint8(2)
	keyLen := uint32(32)
	hash := argon2.IDKey([]byte(payload), salt, t, memory, threads, keyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	hased := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", memory, t, threads, b64Salt, b64Hash)

	return hased, nil
}

func VerifyHash(hashed, plain string) (bool, error) {
	// Hier sollte die eigentliche Hashingsverifizierungs-Logik implementiert werden.
	parts := strings.Split(hashed, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("Ungültiges Hash-Format")
	}

	// Manual parse: m=65536,t=3,p=2
	paramPart := parts[3]
	paramItems := strings.Split(paramPart, ",")
	if len(paramItems) != 3 {
		return false, fmt.Errorf("Ungültiges Parameter im Hash")
	}

	var memory uint32
	var time uint32
	var threads uint8

	for _, item := range paramItems {
		keyVal := strings.Split(item, "=")
		if len(keyVal) != 2 {
			return false, fmt.Errorf("Ungültiges key-value Paar im Parameter")
		}
		switch keyVal[0] {
		case "m":
			mem, err := strconv.ParseUint(keyVal[1], 10, 32)
			if err != nil {
				return false, err
			}
			memory = uint32(mem)
		case "t":
			t, err := strconv.ParseUint(keyVal[1], 10, 32)
			if err != nil {
				return false, err
			}
			time = uint32(t)
		case "p":
			p, err := strconv.ParseUint(keyVal[1], 10, 8)
			if err != nil {
				return false, err
			}
			threads = uint8(p)
		default:
			return false, fmt.Errorf("Unbekannter Parameter: %s", keyVal[0])
		}
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	keyLen := uint32(len(expectedHash))
	computeHash := argon2.IDKey([]byte(plain), salt, time, memory, threads, keyLen)

	if subtle.ConstantTimeCompare(expectedHash, computeHash) == 1 {
		return true, nil
	}

	return false, nil
}
