package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (a *Admin) loginByPassword(w http.ResponseWriter, r *http.Request) {
	var t struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&t); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `invalid body: %v`, err)
		return
	}
	success := a.auth.AuthLogin(t.Username, t.Password)
	if success {
		a.auth.MakeCookie(w, r)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func (a *Admin) loginByGithub(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	success := a.auth.AuthGitHub(code).IsAdmin()
	if success {
		a.auth.MakeCookie(w, r)
		http.Redirect(w, r, `/`, http.StatusFound)
	} else {
		http.Redirect(w, r, a.prefixed(`/login`), http.StatusFound)
	}
}

func (a *Admin) loginByGoogle(w http.ResponseWriter, r *http.Request) {
	var t struct {
		Token string `json:"token"`
	}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&t); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `invalid body: %v`, err)
		return
	}
	success := a.auth.AuthGoogle(t.Token).IsAdmin()
	if success {
		a.auth.MakeCookie(w, r)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}
