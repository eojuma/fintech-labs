package auth

import "errors"

// WebAuthnCredential is the persistence boundary for a future WebAuthn
// ceremony. Raw credential material is never accepted from application forms.
type WebAuthnCredential struct {
	UserID       uint
	CredentialID []byte
	PublicKey    []byte
}

var ErrWebAuthnNotConfigured = errors.New("WebAuthn ceremony provider is not configured")

func BeginRegistration(userID uint) error { return ErrWebAuthnNotConfigured }
