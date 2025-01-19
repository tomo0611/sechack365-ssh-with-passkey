package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"
)

var (
	webAuthn *webauthn.WebAuthn
	err      error

	datastore PasskeyStore
	//sessions  SessionStore
	l Logger
)

type Logger interface {
	Printf(format string, v ...interface{})
}

type PasskeyStore interface {
	GetOrCreateUser(userName string) PasskeyUser
	SaveUser(PasskeyUser)
	GenSessionID() (string, error)
	GetSession(token string) (webauthn.SessionData, bool)
	SaveSession(token string, data webauthn.SessionData)
	DeleteSession(token string)
	GetLoginToken(token string) (LoginToken, bool)
	SaveLoginToken(token LoginToken)
	DeleteLoginToken(token string)
}

var login_approved_tokens []string
var logintoken_session map[string]string = make(map[string]string)

func main() {
	l = log.Default()

	wconfig := &webauthn.Config{
		RPDisplayName: "SSH Login With Passkey (Demo)",
		RPID:          "passkey.tomo0611.dev",
		// The origin URLs allowed for WebAuthn registration/authentication
		RPOrigins: []string{"https://passkey.tomo0611.dev"},
	}

	l.Printf("[INFO] create webauthn")
	if webAuthn, err = webauthn.New(wconfig); err != nil {
		fmt.Printf("[FATA] %s", err.Error())
		os.Exit(1)
	}

	l.Printf("[INFO] create datastore")
	datastore = NewInMem(l)

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasPrefix(c.Request().URL.Path, "/login/") {
				// check /login:loginToken exist in datastore
				loginToken := strings.TrimPrefix(c.Request().URL.Path, "/login/")
				if _, ok := datastore.GetLoginToken(loginToken); !ok {
					return c.String(http.StatusNotFound, "404 NOT FOUND")
				}
			}
			return next(c)
		}
	})
	e.File("/login/:loginToken", "web/login.html")
	e.Static("/", "web")

	/* Login用のTokenを生成 */
	e.GET("/api/v1/generateLoginToken", func(c echo.Context) error {
		username := c.QueryParam("username")
		hostname := c.QueryParam("hostname")
		ip := c.Request().Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = c.RealIP()
		}
		randomString := generateRandomLoginToken()
		datastore.SaveLoginToken(LoginToken{
			Token:    randomString,
			UserName: username,
			HostName: hostname,
			IpAddr:   ip,
			DateTime: time.Now(),
		})
		return c.String(http.StatusOK, randomString)
	})

	/* Loginされたかどうかを確認するAPI */
	e.GET("/api/v1/loginLongPolling/:loginToken", func(c echo.Context) error {
		loginToken := c.Param("loginToken")

		// 1分間だけログインを維持
		for i := 0; i < 60; i++ {
			for _, token := range login_approved_tokens {
				if token == loginToken {
					// delete login token from login_approved_tokens
					deleteLoginApprovedToken(loginToken)
					return c.String(http.StatusOK, "OK")
				}
			}
			time.Sleep(1 * time.Second)
		}
		// 408を返却
		return c.String(http.StatusRequestTimeout, "NG")
	})

	/* 端末のPasskey登録開始に使うAPI */
	e.POST("/api/v1/passkey/registerStart", func(c echo.Context) error {
		user := datastore.GetOrCreateUser("alya") // Find or create the new user
		options, session, err := webAuthn.BeginRegistration(user)
		if err != nil {
			msg := fmt.Sprintf("can't begin registration: %s", err.Error())
			l.Printf("[ERRO] %s", msg)
			return c.String(http.StatusInternalServerError, msg)
		}
		// Make a session key and store the sessionData values
		t, err := datastore.GenSessionID()
		if err != nil {
			l.Printf("[ERRO] can't generate session id: %s", err.Error())

			panic(err) // FIXME: handle error
		}

		datastore.SaveSession(t, *session)

		http.SetCookie(c.Response(), &http.Cookie{
			Name:   "sid",
			Value:  t,
			Path:   "/api/v1/passkey/",
			MaxAge: 3600,
			// For debugging purposes, we can set the cookie to be sent only over HTTPS
			// Cookie “sid” has been rejected because a non-HTTPS cookie can’t be set as “secure”.
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		return c.JSON(http.StatusOK, options)
	})

	/* 端末のPasskey登録終了に使うAPI */
	e.POST("/api/v1/passkey/registerFinish", func(c echo.Context) error {
		// Get the session key from cookie
		sid, err := c.Request().Cookie("sid")
		if err != nil {
			l.Printf("[ERRO] can't get session id: %s", err.Error())
			panic(err)
		}
		// Get the session data stored from the function above
		session, _ := datastore.GetSession(sid.Value)
		user := datastore.GetOrCreateUser("alya")
		credential, err := webAuthn.FinishRegistration(user, session, c.Request())
		if err != nil {
			msg := fmt.Sprintf("can't finish registration: %s", err.Error())
			l.Printf("[ERRO] %s", msg)
			http.SetCookie(c.Response(), &http.Cookie{
				Name:  "sid",
				Value: "",
			})
			return c.JSON(http.StatusBadRequest, msg)
		}

		// If creation was successful, store the credential object
		user.AddCredential(credential)
		datastore.SaveUser(user)
		// Delete the session data
		datastore.DeleteSession(sid.Value)
		http.SetCookie(c.Response(), &http.Cookie{
			Name:  "sid",
			Value: "",
		})
		l.Printf("[INFO] finish registration ----------------------/")
		return c.JSON(http.StatusOK, "Registration Success") // Handle next steps
	})

	e.GET("/api/v1/loginInfo/:loginToken", func(c echo.Context) error {
		loginToken := c.Param("loginToken")
		token, exists := datastore.GetLoginToken(loginToken)
		if !exists {
			return c.JSON(http.StatusBadRequest, "Invalid login token")
		}
		ip := c.Request().Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = c.RealIP()
		}
		token.YourIpAddr = ip
		return c.JSON(http.StatusOK, token)
	})

	e.POST("/api/v1/passkey/loginStart/:loginToken", func(c echo.Context) error {
		loginToken := c.Param("loginToken")
		user := datastore.GetOrCreateUser("alya") // Find the user
		options, session, err := webAuthn.BeginLogin(user)
		if err != nil {
			msg := fmt.Sprintf("can't begin login: %s", err.Error())
			l.Printf("[ERRO] %s", msg)
			return c.JSON(http.StatusBadRequest, msg)
		}

		// Make a session key and store the sessionData values
		t, err := datastore.GenSessionID()
		if err != nil {
			l.Printf("[ERRO] can't generate session id: %s", err.Error())
			panic(err) // TODO: handle error
		}
		datastore.SaveSession(t, *session)
		logintoken_session[loginToken] = t
		return c.JSON(http.StatusOK, options)
	})

	e.POST("/api/v1/passkey/loginFinish/:loginToken", func(c echo.Context) error {
		loginToken := c.Param("loginToken")
		if _, ok := datastore.GetLoginToken(loginToken); !ok {
			return c.JSON(http.StatusBadRequest, "Invalid login token")
		}
		// Get the session key from cookie
		sid := logintoken_session[loginToken]

		// Get the session data stored from the function above
		session, _ := datastore.GetSession(sid) // FIXME: cover invalid session
		l.Printf("[INFO] loginFinish: %v", session)
		// In out example username == userID, but in real world it should be different
		user := datastore.GetOrCreateUser("alya") // Get the user
		l.Printf("[INFO] loginFinish: %v", user)
		credential, err := webAuthn.FinishLogin(user, session, c.Request())
		if err != nil {
			l.Printf("[ERRO] can't finish login: %s", err.Error())
			panic(err)
		}
		// Handle credential.Authenticator.CloneWarning
		if credential.Authenticator.CloneWarning {
			l.Printf("[WARN] can't finish login: %s", "CloneWarning")
		}
		// If login was successful, update the credential object
		user.UpdateCredential(credential)
		datastore.SaveUser(user)
		// Delete the login session data
		datastore.DeleteSession(sid)
		// LoginTokenを削除
		datastore.DeleteLoginToken(loginToken)
		// Add login token to login_approved_tokens
		login_approved_tokens = append(login_approved_tokens, loginToken)

		l.Printf("[INFO] finish login ----------------------/")
		return c.JSON(http.StatusOK, "Login Success")
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// Start server
	go func() {
		if err := e.Start("127.0.0.1:1323"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatalf("shutting down the server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

}

func deleteLoginApprovedToken(token string) {
	for i, t := range login_approved_tokens {
		if t == token {
			login_approved_tokens = append(login_approved_tokens[:i], login_approved_tokens[i+1:]...)
			break
		}
	}
}
