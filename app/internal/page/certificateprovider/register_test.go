package certificateprovider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Print(v ...any) {
	m.Called(v...)
}

type mockSessionsStore struct {
	mock.Mock
}

func (m *mockSessionsStore) New(r *http.Request, name string) (*sessions.Session, error) {
	args := m.Called(r, name)
	return args.Get(0).(*sessions.Session), args.Error(1)
}

func (m *mockSessionsStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	args := m.Called(r, name)
	return args.Get(0).(*sessions.Session), args.Error(1)
}

func (m *mockSessionsStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	args := m.Called(r, w, session)
	return args.Error(0)
}

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)
	localizer := localize.Localizer{}

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{
			Values: map[any]any{
				"certificate-provider": &sesh.CertificateProviderSession{
					Sub:            "random",
					DonorSessionID: "session-id",
					LpaID:          "lpa-id",
				},
			},
		}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionsStore, localizer, localize.En, page.RumConfig{ApplicationID: "xyz"}, "?%3fNEI0t9MN", None)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:             "/path",
			Query:            "?a=b",
			Localizer:        localizer,
			Lang:             localize.En,
			SessionID:        "session-id",
			LpaID:            "lpa-id",
			CookieConsentSet: false,
			CanGoBack:        false,
			RumConfig:        page.RumConfig{ApplicationID: "xyz"},
			StaticHash:       "?%3fNEI0t9MN",
			Paths:            page.Paths,
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, r.WithContext(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: "session-id", LpaID: "lpa-id"})), hr)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestMakeHandleExistingSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "ignored-123", SessionID: "ignored-session-id"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/path?a=b", nil)
	localizer := localize.Localizer{}

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"certificate-provider": &sesh.CertificateProviderSession{Sub: "random", LpaID: "lpa-id", DonorSessionID: "session-id"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionsStore, localizer, localize.En, page.RumConfig{ApplicationID: "xyz"}, "?%3fNEI0t9MN", None)
	handle("/path", RequireSession|CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:             "/path",
			Query:            "?a=b",
			Localizer:        localizer,
			Lang:             localize.En,
			SessionID:        "session-id",
			CookieConsentSet: false,
			CanGoBack:        true,
			RumConfig:        page.RumConfig{ApplicationID: "xyz"},
			StaticHash:       "?%3fNEI0t9MN",
			Paths:            page.Paths,
			LpaID:            "lpa-id",
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, r.WithContext(page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id", SessionID: "session-id"})), hr)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestMakeHandleShowTranslationKeys(t *testing.T) {
	testCases := map[string]struct {
		showTranslationKeys string
		expected            bool
	}{
		"requested": {
			showTranslationKeys: "1",
			expected:            true,
		},
		"not requested": {
			showTranslationKeys: "maybe",
			expected:            false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/path?showTranslationKeys="+tc.showTranslationKeys, nil)
			localizer := localize.Localizer{}

			mux := http.NewServeMux()
			handle := makeHandle(mux, nil, nil, localizer, localize.En, page.RumConfig{ApplicationID: "xyz"}, "?%3fNEI0t9MN", None)
			handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
				expectedLocalizer := localize.Localizer{}
				expectedLocalizer.ShowTranslationKeys = tc.expected

				assert.Equal(t, page.AppData{
					Page:             "/path",
					Query:            "?showTranslationKeys=" + tc.showTranslationKeys,
					Localizer:        expectedLocalizer,
					Lang:             localize.En,
					CookieConsentSet: false,
					RumConfig:        page.RumConfig{ApplicationID: "xyz"},
					StaticHash:       "?%3fNEI0t9MN",
					Paths:            page.Paths,
				}, appData)
				assert.Equal(t, w, hw)
				hw.WriteHeader(http.StatusTeapot)
				return nil
			})

			mux.ServeHTTP(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusTeapot, resp.StatusCode)
		})
	}
}

func TestMakeHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	localizer := localize.Localizer{}

	logger := &mockLogger{}
	logger.
		On("Print", fmt.Sprintf("Error rendering page for path '%s': %s", "/path", expectedError.Error()))

	mux := http.NewServeMux()
	handle := makeHandle(mux, logger, nil, localizer, localize.En, page.RumConfig{}, "?%3fNEI0t9MN", None)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestMakeHandleSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	localizer := localize.Localizer{}

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{}, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, logger, sessionsStore, localizer, localize.En, page.RumConfig{}, "?%3fNEI0t9MN", None)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore, logger)
}

func TestMakeHandleSessionMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	localizer := localize.Localizer{}

	logger := &mockLogger{}
	logger.
		On("Print", sesh.MissingSessionError("certificate-provider"))

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, logger, sessionsStore, localizer, localize.En, page.RumConfig{}, "?%3fNEI0t9MN", None)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore, logger)
}

func TestMakeHandleNoSessionRequired(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	localizer := localize.Localizer{}

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil, localizer, localize.En, page.RumConfig{}, "?%3fNEI0t9MN", None)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:             "/path",
			Localizer:        localizer,
			Lang:             localize.En,
			CookieConsentSet: false,
			StaticHash:       "?%3fNEI0t9MN",
			Paths:            page.Paths,
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, r, hr)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}
