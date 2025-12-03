package auth_webauthn

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/movsb/taoblog/modules/auth/cookies"
	"github.com/movsb/taoblog/modules/auth/user"
	"github.com/phuslu/lru"
)

type WebAuthnUser struct {
	*user.User
	credentials []webauthn.Credential
}

func (u *WebAuthnUser) WebAuthnID() []byte {
	if u.ID <= 0 || u.ID > math.MaxInt32 {
		panic(`user id is invalid`)
	}
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(u.ID))
	return buf
}
func (u *WebAuthnUser) WebAuthnName() string {
	return fmt.Sprintf(`%s (id:%d)`, u.Nickname, u.ID)
}
func (u *WebAuthnUser) WebAuthnDisplayName() string {
	return u.Nickname
}
func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}
func (u *WebAuthnUser) WebAuthnIcon() string {
	return ""
}

func ToWebAuthnUser(u *user.User) webauthn.User {
	return &WebAuthnUser{
		User:        u,
		credentials: u.Credentials,
	}
}

type AuthBackend interface {
	GetUserByID(ctx context.Context, id int) (*user.User, error)
	AddWebAuthnCredential(user *user.User, credential *webauthn.Credential)
}

// 基于 webauthn 代码库提供前端的 webauthn 登录接口。
type WebAuthn struct {
	wa *webauthn.WebAuthn

	authBackend AuthBackend

	// 注册/登录过程中临时保存的会话信息。
	// TOOD 换成 file cache，以不限制容量。
	registrationSessions *lru.TTLCache[int64, *webauthn.SessionData]
	loginSessions        *lru.TTLCache[string, *webauthn.SessionData]
}

const webAuthnSessionTTL = time.Minute * 5

func NewWebAuthn(authBackend AuthBackend, getWebAuthn *webauthn.WebAuthn) *WebAuthn {
	return &WebAuthn{
		authBackend: authBackend,
		wa:          getWebAuthn,

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

	mux.HandleFunc(`POST /register:begin`, a.beginRegistration)
	mux.HandleFunc(`POST /register:finish`, a.finishRegistration)
	mux.HandleFunc(`POST /login:begin`, a.beginLogin)
	mux.HandleFunc(`POST /login:finish`, a.finishLogin)

	return mux
}

func (a *WebAuthn) beginRegistration(w http.ResponseWriter, r *http.Request) {
	user := user.MustNotBeGuest(r.Context()).User

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

func (a *WebAuthn) finishRegistration(w http.ResponseWriter, r *http.Request) {
	user := user.MustNotBeGuest(r.Context()).User

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

	a.authBackend.AddWebAuthnCredential(user, credential)
	log.Println("注册成功，结果已保存")

	w.WriteHeader(http.StatusOK)
}

func (a *WebAuthn) beginLogin(w http.ResponseWriter, r *http.Request) {
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

func (a *WebAuthn) finishLogin(w http.ResponseWriter, r *http.Request) {
	challenge := r.URL.Query().Get(`challenge`)
	session, ok := a.loginSessions.Get(challenge)
	if !ok {
		http.Error(w, "会话不存在或已经过期。", http.StatusBadRequest)
		return
	}
	defer a.loginSessions.Delete(challenge)

	var outUser *user.User

	credential, err := a.wa.FinishDiscoverableLogin(
		func(rawID, userHandle []byte) (webauthn.User, error) {
			id := binary.LittleEndian.Uint32(userHandle)
			u, err := a.authBackend.GetUserByID(r.Context(), int(id))
			if err != nil {
				return nil, err
			}
			outUser = u
			return ToWebAuthnUser(u), nil
		},
		*session, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a.authBackend.AddWebAuthnCredential(outUser, credential)
	log.Println("登录成功，凭证已更新")

	cookies.MakeCookie(w, r, int(outUser.ID), outUser.Password, outUser.Nickname)

	w.WriteHeader(http.StatusOK)
}
