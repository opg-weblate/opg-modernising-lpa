package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityWithYotiCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)
	now := time.Now()
	userData := identity.UserData{FullName: "a-full-name", RetrievedAt: now}

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)
	lpaStore.On("Put", r.Context(), &page.Lpa{YotiUserData: userData}).Return(nil)

	yotiClient := &mockYotiClient{}
	yotiClient.On("User", "a-token").Return(userData, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &identityWithYotiCallbackData{
			App:         testAppData,
			FullName:    "a-full-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithYotiCallback(template.Func, yotiClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, yotiClient, template)
}

func TestGetIdentityWithYotiCallbackWhenError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)

	yotiClient := &mockYotiClient{}
	yotiClient.On("User", "a-token").Return(identity.UserData{}, expectedError)

	err := IdentityWithYotiCallback(nil, yotiClient, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, yotiClient)
}

func TestGetIdentityWithYotiCallbackWhenGetDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, expectedError)

	err := IdentityWithYotiCallback(nil, nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetIdentityWithYotiCallbackWhenPutDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)
	now := time.Now()
	userData := identity.UserData{FullName: "a-full-name", RetrievedAt: now}

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)
	lpaStore.On("Put", r.Context(), &page.Lpa{YotiUserData: userData}).Return(expectedError)

	yotiClient := &mockYotiClient{}
	yotiClient.On("User", "a-token").Return(userData, nil)

	err := IdentityWithYotiCallback(nil, yotiClient, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, yotiClient)
}

func TestGetIdentityWithYotiCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{OK: true, FullName: "a-full-name", RetrievedAt: now}

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{YotiUserData: userData}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &identityWithYotiCallbackData{
			App:         testAppData,
			FullName:    "a-full-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithYotiCallback(template.Func, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostIdentityWithYotiCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{
		IdentityOption: identity.EasyID,
	}, nil)

	err := IdentityWithYotiCallback(nil, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ReadYourLpa, resp.Header.Get("Location"))
}
