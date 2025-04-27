package auth

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/movsb/taoblog/service/models"
)

type WebAuthn struct {
	wa   *webauthn.WebAuthn
	auth *Auth

	// 注册/登录过程中临时保存的会话信息。
	// map[user_id_int64]*webauthn.SessionData
	registrationSessions sync.Map
	loginSessions        sync.Map
}

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

func (a *WebAuthn) Handler(prefix string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(`POST /register:begin`, a.BeginRegistration)
	mux.HandleFunc(`POST /register:finish`, a.FinishRegistration)
	mux.HandleFunc(`POST /login:begin`, a.BeginLogin)
	mux.HandleFunc(`POST /login:finish`, a.FinishLogin)

	return http.StripPrefix(strings.TrimSuffix(prefix, "/"), mux)
}

func (a *WebAuthn) BeginRegistration(w http.ResponseWriter, r *http.Request) {
	user := a.auth.AuthRequest(r)
	if user.IsGuest() {
		// 暂时不允许注册，因为没有用户数据库。管理员是存在选项表中的🥵。
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

	a.registrationSessions.Store(user.ID, session)
}

func (a *WebAuthn) FinishRegistration(w http.ResponseWriter, r *http.Request) {
	user := a.auth.AuthRequest(r)
	if user.IsGuest() {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	sessionAny, ok := a.registrationSessions.Load(user.ID)
	if !ok {
		http.Error(w, "会话不存在", http.StatusBadRequest)
		return
	}
	session := sessionAny.(*webauthn.SessionData)
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
	a.loginSessions.Store(session.Challenge, session)
}

func (a *WebAuthn) FinishLogin(w http.ResponseWriter, r *http.Request) {
	challenge := r.URL.Query().Get(`challenge`)
	sessionAny, ok := a.loginSessions.Load(challenge)
	if !ok {
		http.Error(w, "会话不存在", http.StatusBadRequest)
		return
	}
	session := sessionAny.(*webauthn.SessionData)
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

	a.auth.MakeCookie(outUser, w, r)

	w.WriteHeader(http.StatusOK)
}
