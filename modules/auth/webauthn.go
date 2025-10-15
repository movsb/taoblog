package auth

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/movsb/taoblog/modules/auth/cookies"
	"github.com/movsb/taoblog/service/models"
	"github.com/phuslu/lru"
)

type WebAuthn struct {
	wa   *webauthn.WebAuthn
	auth *Auth

	// 注册/登录过程中临时保存的会话信息。
	// TOOD 换成 file cache，以不限制容量。
	registrationSessions *lru.TTLCache[int64, *webauthn.SessionData]
	loginSessions        *lru.TTLCache[string, *webauthn.SessionData]
}

const webAuthnSessionTTL = time.Minute * 5

func NewWebAuthn(auth *Auth, domain string, displayName string, origins []string) *WebAuthn {
	config := &webauthn.Config{
		RPID:          domain,
		RPDisplayName: displayName,
		RPOrigins:     origins,
	}
	wa, err := webauthn.New(config)
	if err != nil {
		panic(err)
	}
	return &WebAuthn{
		auth: auth,
		wa:   wa,

		registrationSessions: lru.NewTTLCache[int64, *webauthn.SessionData](8),
		loginSessions:        lru.NewTTLCache[string, *webauthn.SessionData](8),
	}
}

func writeJsonBody(w http.ResponseWriter, data any) error {
	w.Header().Add(`Content-Type`, `application/json`)
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(data); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (a *WebAuthn) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(`POST /register:begin`, a.BeginRegistration)
	mux.HandleFunc(`POST /register:finish`, a.FinishRegistration)
	mux.HandleFunc(`POST /login:begin`, a.BeginLogin)
	mux.HandleFunc(`POST /login:finish`, a.FinishLogin)

	return mux
}

func (a *WebAuthn) BeginRegistration(w http.ResponseWriter, r *http.Request) {
	user := a.auth.authRequest(w, r)
	if user.IsGuest() {
		http.Error(w, "不允许非管理员用户注册通行密钥。", http.StatusForbidden)
		return
	}

	options, session, err := a.wa.BeginRegistration(ToWebAuthnUser(user),
		// YubiKey + Webauth: userHandle is always null
		// https://stackoverflow.com/a/62780333/3628322
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := writeJsonBody(w, options); err != nil {
		return
	}

	a.registrationSessions.Set(user.ID, session, webAuthnSessionTTL)
}

func (a *WebAuthn) FinishRegistration(w http.ResponseWriter, r *http.Request) {
	user := a.auth.authRequest(w, r)
	if user.IsGuest() {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	session, ok := a.registrationSessions.Get(user.ID)
	if !ok {
		http.Error(w, "会话不存在或已经过期。", http.StatusBadRequest)
		return
	}
	defer a.registrationSessions.Delete(user.ID)

	credential, err := a.wa.FinishRegistration(ToWebAuthnUser(user), *session, r)
	if err != nil {
		http.Error(w, "注册失败："+err.Error(), http.StatusInternalServerError)
		return
	}

	a.auth.AddWebAuthnCredential(user, credential)
	log.Println("注册成功，结果已保存")

	w.WriteHeader(http.StatusOK)
}

func (a *WebAuthn) BeginLogin(w http.ResponseWriter, r *http.Request) {
	options, session, err := a.wa.BeginDiscoverableLogin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := writeJsonBody(w, options); err != nil {
		return
	}
	a.loginSessions.Set(session.Challenge, session, webAuthnSessionTTL)
}

func (a *WebAuthn) FinishLogin(w http.ResponseWriter, r *http.Request) {
	challenge := r.URL.Query().Get(`challenge`)
	session, ok := a.loginSessions.Get(challenge)
	if !ok {
		http.Error(w, "会话不存在或已经过期。", http.StatusBadRequest)
		return
	}
	defer a.loginSessions.Delete(challenge)

	var outUser *User

	credential, err := a.wa.FinishDiscoverableLogin(
		func(rawID, userHandle []byte) (webauthn.User, error) {
			id := binary.LittleEndian.Uint32(userHandle)
			var user models.User
			if err := a.auth.db.Where(`id=?`, id).Find(&user); err != nil {
				return nil, err
			}
			outUser = &User{&user}
			return ToWebAuthnUser(outUser), nil
		},
		*session, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a.auth.AddWebAuthnCredential(outUser, credential)
	log.Println("登录成功，凭证已更新")

	cookies.MakeCookie(w, r, int(outUser.ID), outUser.Password, outUser.Nickname)

	w.WriteHeader(http.StatusOK)
}
