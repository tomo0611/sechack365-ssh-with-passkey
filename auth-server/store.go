package main

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/go-webauthn/webauthn/webauthn"
)

type InMem struct {
	// usernameとPasskeyUserのmap
	users map[string]PasskeyUser
	// sessionIDとSessionDataのmap
	sessions map[string]webauthn.SessionData
	// LoginTokens
	loginTokens []LoginToken
	log         Logger
}

func (i *InMem) GenSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func NewInMem(log Logger) *InMem {
	return &InMem{
		users:       make(map[string]PasskeyUser),
		sessions:    make(map[string]webauthn.SessionData),
		loginTokens: []LoginToken{},
		log:         log,
	}
}

func (i *InMem) GetSession(token string) (webauthn.SessionData, bool) {
	val, ok := i.sessions[token]
	return val, ok
}

func (i *InMem) SaveSession(token string, data webauthn.SessionData) {
	i.sessions[token] = data
}

func (i *InMem) DeleteSession(token string) {
	delete(i.sessions, token)
}

func (i *InMem) GetOrCreateUser(userName string) PasskeyUser {
	i.log.Printf("[DEBUG] GetOrCreateUser: %v", userName)
	if _, ok := i.users[userName]; !ok {
		i.log.Printf("[DEBUG] GetOrCreateUser: creating new user: %v", userName)
		i.users[userName] = &User{
			ID:          []byte(userName),
			DisplayName: "Alisa Mikhailovna Kujou",
			Name:        userName,
		}
	}

	return i.users[userName]
}

func (i *InMem) SaveUser(user PasskeyUser) {
	i.users[user.WebAuthnName()] = user
}

func (i *InMem) GetLoginToken(token string) (LoginToken, bool) {
	for _, t := range i.loginTokens {
		if t.Token == token {
			return t, true
		}
	}
	return LoginToken{}, false
}

func (i *InMem) SaveLoginToken(token LoginToken) {
	i.loginTokens = append(i.loginTokens, token)
}

func (mem *InMem) DeleteLoginToken(token string) {
	for i, t := range mem.loginTokens {
		if t.Token == token {
			mem.loginTokens = append(mem.loginTokens[:i], mem.loginTokens[i+1:]...)
			return
		}
	}
}
