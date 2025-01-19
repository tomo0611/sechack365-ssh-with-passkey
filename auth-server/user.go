package main

import "github.com/go-webauthn/webauthn/webauthn"

// Userはwebauthn.Userを実装している
type PasskeyUser interface {
	webauthn.User
	AddCredential(*webauthn.Credential)
	UpdateCredential(*webauthn.Credential)
}

type User struct {
	ID          []byte
	DisplayName string
	Name        string
	// 1ユーザーに対して複数のCredentialを持たせるためのスライス
	creds []webauthn.Credential
}

// WebAuthnライブラリを使うには以下の4関数の実装が必須
// 以下を参照するとわかりやすい
// https://github.com/go-webauthn/webauthn/blob/master/webauthn/types.go#L173
func (o *User) WebAuthnID() []byte {
	return o.ID
}

func (o *User) WebAuthnName() string {
	return o.Name
}

func (o *User) WebAuthnDisplayName() string {
	return o.DisplayName
}

func (o *User) WebAuthnCredentials() []webauthn.Credential {
	return o.creds
}

// 逆にここは任意、今回はCredentialの追加と更新を行う関数を追加している
func (o *User) AddCredential(credential *webauthn.Credential) {
	o.creds = append(o.creds, *credential)
}

// IDが一致するCredentialを更新する
func (o *User) UpdateCredential(credential *webauthn.Credential) {
	for i, c := range o.creds {
		if string(c.ID) == string(credential.ID) {
			o.creds[i] = *credential
		}
	}
}
