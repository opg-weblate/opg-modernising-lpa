package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockYotiClient struct {
	mock.Mock
}

func (m *mockYotiClient) IsTest() bool {
	return m.Called().Bool(0)
}

func (m *mockYotiClient) SdkID() string {
	return m.Called().String(0)
}

func (m *mockYotiClient) User(token string) (identity.UserData, error) {
	args := m.Called(token)

	return args.Get(0).(identity.UserData), args.Error(1)
}

func TestGetIdentityWithYoti(t *testing.T) {
	w := httptest.NewRecorder()

	yotiClient := &mockYotiClient{}
	yotiClient.On("IsTest").Return(false)
	yotiClient.On("SdkID").Return("an-sdk-id")

	template := &mockTemplate{}
	template.
		On("Func", w, &identityWithYotiData{
			App:         appData,
			ClientSdkID: "an-sdk-id",
			ScenarioID:  "a-scenario-id",
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := IdentityWithYoti(template.Func, yotiClient, "a-scenario-id")(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, yotiClient, template)
}

func TestGetIdentityWithYotiWhenTest(t *testing.T) {
	w := httptest.NewRecorder()

	yotiClient := &mockYotiClient{}
	yotiClient.On("IsTest").Return(true)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := IdentityWithYoti(nil, yotiClient, "")(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, identityWithYotiCallbackPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, yotiClient)
}

func TestGetIdentityWithYotiWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()

	yotiClient := &mockYotiClient{}
	yotiClient.On("IsTest").Return(false)
	yotiClient.On("SdkID").Return("an-sdk-id")

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := IdentityWithYoti(template.Func, yotiClient, "a-scenario-id")(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, yotiClient, template)
}