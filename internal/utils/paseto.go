package utils

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"aidanwoods.dev/go-paseto"
)

// PasetoMaker verarbeitet lokale PASETO-Operationen der Version 4 (symmetrisch).
type PasetoMaker struct {
	symmetricKey paseto.V4SymmetricKey
}

// NewPasetoMaker creates instance with existing key
func NewPasetoMaker(keyHex string) (*PasetoMaker, error) {
	key, err := paseto.V4SymmetricKeyFromHex(keyHex)
	if err != nil {
		return nil, fmt.Errorf("Invalid symmetric key: %w", err)
	}

	return &PasetoMaker{
		symmetricKey: key,
	}, nil
}

// GenerateSymmetricKey generiert einen neuen symmetrischen V4-Schlüssel. Wird verwendet, wenn kein hexKey vorhanden ist, nur einmal.
func GenerateSymmetricKey() string {
	key := paseto.NewV4SymmetricKey()
	return hex.EncodeToString(key.ExportBytes())
}

// CreateToken erstellt ein lokales V4 Token (encrypted)
func (m *PasetoMaker) CreateToken(userID, username, email, sessionID string, isUserActive bool, duration time.Duration) (string, error) {
	token := paseto.NewToken()

	// Standard Claims festlegen
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(duration))
	token.SetAudience("Aufgaben-meister")
	token.SetIssuer("AM-service")
	token.SetSubject(userID)

	// Benutzerdefiniert Claims festlegen
	token.SetString("username", username)
	token.SetString("email", email)
	token.SetString("is_active", strconv.FormatBool(isUserActive))
	token.SetString("jti", sessionID)

	// Encrypt mit V4 local (symmetric)
	encypted := token.V4Encrypt(m.symmetricKey, nil)

	return encypted, nil
}

type PayloadPaseto struct {
	UserID   string
	Username string
	Email    string
	IsActive string
	JTI      string
	Duration time.Time
}

// VerifyToken decrypts und überprüft das lokale V4 Token.
func (m *PasetoMaker) VerifyToken(tokenString string) (*PayloadPaseto, error) {
	parser := paseto.NewParser()

	// Validierungsregeln hinzufügen
	parser.AddRule(paseto.NotExpired())
	parser.AddRule(paseto.ForAudience("Aufgaben-meister"))
	parser.AddRule(paseto.ValidAt(time.Now()))

	// Parse und decrypt mit symmetrischem Schlüssel
	parsedToken, err := parser.ParseV4Local(m.symmetricKey, tokenString, nil)
	if err != nil {
		return nil, fmt.Errorf("Token decryption/verification failed: %w", err)
	}

	claims := parsedToken.Claims()

	userID, _ := claims["sub"].(string)
	username, _ := claims["username"].(string)
	email, _ := claims["email"].(string)
	isActive, _ := claims["is_Active"].(string)
	JTI, _ := claims["jti"].(string)

	var exp time.Time
	if t, ok := claims["exp"].(time.Time); ok {
		exp = t
	} else if s, ok := claims["exp"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, s); err == nil {
			exp = parsed
		}
	} else if f, ok := claims["exp"].(float64); ok {
		exp = time.Unix(int64(f), 0)
	}

	payload := &PayloadPaseto{
		UserID:   userID,
		Username: username,
		Email:    email,
		IsActive: isActive,
		JTI:      JTI,
		Duration: exp,
	}

	return payload, nil

}
