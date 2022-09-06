package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWhenCanTheLpaBeUsed(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App: appData,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WhenCanTheLpaBeUsed(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestGetWhenCanTheLpaBeUsedFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{data: Lpa{WhenCanTheLpaBeUsed: "when-registered"}}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App:  appData,
			When: "when-registered",
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WhenCanTheLpaBeUsed(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestGetWhenCanTheLpaBeUsedWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WhenCanTheLpaBeUsed(nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestGetWhenCanTheLpaBeUsedWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App: appData,
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WhenCanTheLpaBeUsed(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestPostWhenCanTheLpaBeUsed(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{WhenCanTheLpaBeUsed: "when-registered"}).
		Return(nil)

	form := url.Values{
		"when": {"when-registered"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := WhenCanTheLpaBeUsed(nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, taskListPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostWhenCanTheLpaBeUsedWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{WhenCanTheLpaBeUsed: "when-registered"}).
		Return(expectedError)

	form := url.Values{
		"when": {"when-registered"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := WhenCanTheLpaBeUsed(nil, dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostWhenCanTheLpaBeUsedWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App: appData,
			Errors: map[string]string{
				"when": "selectWhenCanTheLpaBeUsed",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := WhenCanTheLpaBeUsed(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadWhenCanTheLpaBeUsedForm(t *testing.T) {
	form := url.Values{
		"when": {"when-registered"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readWhenCanTheLpaBeUsedForm(r)

	assert.Equal(t, "when-registered", result.When)
}

func TestWhenCanTheLpaBeUsedFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *whenCanTheLpaBeUsedForm
		errors map[string]string
	}{
		"when-registered": {
			form: &whenCanTheLpaBeUsedForm{
				When: "when-registered",
			},
			errors: map[string]string{},
		},
		"when-capacity-lost": {
			form: &whenCanTheLpaBeUsedForm{
				When: "when-capacity-lost",
			},
			errors: map[string]string{},
		},
		"missing": {
			form: &whenCanTheLpaBeUsedForm{},
			errors: map[string]string{
				"when": "selectWhenCanTheLpaBeUsed",
			},
		},
		"invalid": {
			form: &whenCanTheLpaBeUsedForm{
				When: "what",
			},
			errors: map[string]string{
				"when": "selectWhenCanTheLpaBeUsed",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
