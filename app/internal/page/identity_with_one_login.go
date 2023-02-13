package page

import (
	"net/http"

	"github.com/gorilla/sessions"
)

func IdentityWithOneLogin(logger Logger, oneLoginClient OneLoginClient, store sessions.Store, randomString func(int) string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		locale := ""
		if appData.Lang == Cy {
			locale = "cy"
		}

		state := randomString(12)
		nonce := randomString(12)

		authCodeURL := oneLoginClient.AuthCodeURL(state, nonce, locale, true)

		if err := setOneLoginSession(store, r, w, &OneLoginSession{
			State:    state,
			Nonce:    nonce,
			Locale:   locale,
			Identity: true,
			LpaID:    appData.LpaID,
		}); err != nil {
			logger.Print(err)
			return nil
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
		return nil
	}
}
