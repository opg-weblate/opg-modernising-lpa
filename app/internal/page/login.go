package page

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type loginClient interface {
	AuthCodeURL(state, nonce string) string
}

func Login(logger Logger, c loginClient, store sessions.Store, randomString func(int) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := randomString(12)
		nonce := randomString(12)

		authCodeURL := c.AuthCodeURL(state, nonce)

		session := sessions.NewSession(store, "params")
		session.Values = map[interface{}]interface{}{
			"state": state,
			"nonce": nonce,
		}
		session.Options.MaxAge = 10 * 60 * 60
		session.Options.SameSite = http.SameSiteStrictMode
		session.Options.HttpOnly = true
		session.Options.Secure = r.URL.Scheme == "https"

		if err := store.Save(r, w, session); err != nil {
			logger.Print(err)
			return
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
	}
}
