package donor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSelectYourIdentityOptions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &selectYourIdentityOptionsData{
			App:  page.TestAppData,
			Page: 2,
			Form: &selectYourIdentityOptionsForm{},
		}).
		Return(nil)

	err := SelectYourIdentityOptions(template.Func, lpaStore, 2)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetSelectYourIdentityOptionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, page.ExpectedError)

	err := SelectYourIdentityOptions(nil, lpaStore, 0)(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetSelectYourIdentityOptionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			IdentityOption: identity.Passport,
		}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &selectYourIdentityOptionsData{
			App:  page.TestAppData,
			Form: &selectYourIdentityOptionsForm{Selected: identity.Passport},
		}).
		Return(nil)

	err := SelectYourIdentityOptions(template.Func, lpaStore, 0)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetSelectYourIdentityOptionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(page.ExpectedError)

	err := SelectYourIdentityOptions(template.Func, lpaStore, 0)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostSelectYourIdentityOptions(t *testing.T) {
	f := url.Values{
		"option": {"passport"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			IdentityOption: identity.Passport,
			Tasks: page.Tasks{
				ConfirmYourIdentityAndSign: page.TaskInProgress,
			},
		}).
		Return(nil)

	err := SelectYourIdentityOptions(nil, lpaStore, 0)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.YourChosenIdentityOptions, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostSelectYourIdentityOptionsNone(t *testing.T) {
	for pageIndex, nextPath := range map[int]string{
		0: "/lpa/lpa-id" + page.Paths.SelectYourIdentityOptions1,
		1: "/lpa/lpa-id" + page.Paths.SelectYourIdentityOptions2,
		2: "/lpa/lpa-id" + page.Paths.TaskList,
	} {
		t.Run(fmt.Sprintf("Page%d", pageIndex), func(t *testing.T) {
			f := url.Values{
				"option": {"none"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := &page.MockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{}, nil)

			err := SelectYourIdentityOptions(nil, lpaStore, pageIndex)(page.TestAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, nextPath, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostSelectYourIdentityOptionsWhenStoreErrors(t *testing.T) {
	f := url.Values{
		"option": {"passport"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(page.ExpectedError)

	err := SelectYourIdentityOptions(nil, lpaStore, 0)(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostSelectYourIdentityOptionsWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &selectYourIdentityOptionsData{
			App:    page.TestAppData,
			Form:   &selectYourIdentityOptionsForm{},
			Errors: validation.With("option", validation.SelectError{Label: "fromTheListedOptions"}),
		}).
		Return(nil)

	err := SelectYourIdentityOptions(template.Func, lpaStore, 0)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadSelectYourIdentityOptionsForm(t *testing.T) {
	f := url.Values{
		"option": {"passport"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readSelectYourIdentityOptionsForm(r)

	assert.Equal(t, identity.Passport, result.Selected)
	assert.False(t, result.None)
}

func TestSelectYourIdentityOptionsFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *selectYourIdentityOptionsForm
		errors validation.List
	}{
		"valid": {
			form: &selectYourIdentityOptionsForm{
				Selected: identity.EasyID,
			},
		},
		"none": {
			form: &selectYourIdentityOptionsForm{
				Selected: identity.UnknownOption,
				None:     true,
			},
		},
		"missing": {
			form:   &selectYourIdentityOptionsForm{},
			errors: validation.With("option", validation.SelectError{Label: "fromTheListedOptions"}),
		},
		"invalid": {
			form: &selectYourIdentityOptionsForm{
				Selected: identity.UnknownOption,
			},
			errors: validation.With("option", validation.SelectError{Label: "fromTheListedOptions"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
