package models

import "github.com/go-webauthn/webauthn/webauthn"

type PasskeyUser struct {
	ID          []byte
	DisplayName string
	Name        string

	Credentials []webauthn.Credential
}

func (u *PasskeyUser) WebAuthnID() []byte {
	return u.ID
}

func (u *PasskeyUser) WebAuthnName() string {
	return u.Name
}

func (u *PasskeyUser) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u *PasskeyUser) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

func (u PasskeyUser) WebAuthnIcon() string {
	return ""
}

func (u *PasskeyUser) PutCredential(credential webauthn.Credential) {
	u.Credentials = append(u.Credentials, credential)
}

func (u *PasskeyUser) AddCredential(credential *webauthn.Credential) {
	u.Credentials = append(u.Credentials, *credential)
}

func (u *PasskeyUser) UpdateCredential(credential *webauthn.Credential) {
	for i, c := range u.Credentials {
		if string(c.ID) == string(credential.ID) {
			u.Credentials[i] = *credential
		}
	}
}
