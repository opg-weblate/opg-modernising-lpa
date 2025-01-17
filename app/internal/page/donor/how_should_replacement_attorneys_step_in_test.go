package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowShouldReplacementAttorneysStepIn(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldReplacementAttorneysStepInData{
			App:  appData,
			Form: &howShouldReplacementAttorneysStepInForm{},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowShouldReplacementAttorneysStepInFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			HowShouldReplacementAttorneysStepIn:        page.SomeOtherWay,
			HowShouldReplacementAttorneysStepInDetails: "some details",
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldReplacementAttorneysStepInData{
			App: appData,
			Form: &howShouldReplacementAttorneysStepInForm{
				WhenToStepIn: page.SomeOtherWay,
				OtherDetails: "some details",
			},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowShouldReplacementAttorneysStepInWhenStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	template := &mockTemplate{}

	err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostHowShouldReplacementAttorneysStepIn(t *testing.T) {
	form := url.Values{
		"when-to-step-in": {page.SomeOtherWay},
		"other-details":   {"some details"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			HowShouldReplacementAttorneysStepIn:        "",
			HowShouldReplacementAttorneysStepInDetails: "",
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			HowShouldReplacementAttorneysStepIn:        page.SomeOtherWay,
			HowShouldReplacementAttorneysStepInDetails: "some details"}).
		Return(nil)

	template := &mockTemplate{}

	err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.TaskList, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostHowShouldReplacementAttorneysStepInRedirects(t *testing.T) {
	testCases := map[string]struct {
		Attorneys                           actor.Attorneys
		ReplacementAttorneys                actor.Attorneys
		HowAttorneysMakeDecisions           string
		HowShouldReplacementAttorneysStepIn string
		ExpectedRedirectUrl                 string
	}{
		"multiple attorneys acting jointly and severally replacements step in when none left": {
			Attorneys: actor.Attorneys{
				{ID: "123"},
				{ID: "123"},
			},
			ReplacementAttorneys: actor.Attorneys{
				{ID: "123"},
				{ID: "123"},
			},
			HowAttorneysMakeDecisions:           "jointly-and-severally",
			HowShouldReplacementAttorneysStepIn: page.AllCanNoLongerAct,
			ExpectedRedirectUrl:                 "/lpa/lpa-id" + page.Paths.HowShouldReplacementAttorneysMakeDecisions,
		},
		"multiple attorneys acting jointly and severally replacements step in when one loses capacity": {
			Attorneys: actor.Attorneys{
				{ID: "123"},
				{ID: "123"},
			},
			HowAttorneysMakeDecisions:           "jointly-and-severally",
			HowShouldReplacementAttorneysStepIn: page.OneCanNoLongerAct,
			ExpectedRedirectUrl:                 "/lpa/lpa-id" + page.Paths.TaskList,
		},
		"multiple attorneys acting jointly": {
			Attorneys: actor.Attorneys{
				{ID: "123"},
				{ID: "123"},
			},
			HowAttorneysMakeDecisions:           "jointly-and-severally",
			HowShouldReplacementAttorneysStepIn: page.OneCanNoLongerAct,
			ExpectedRedirectUrl:                 "/lpa/lpa-id" + page.Paths.TaskList,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"when-to-step-in": {tc.HowShouldReplacementAttorneysStepIn},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					HowAttorneysMakeDecisions: tc.HowAttorneysMakeDecisions,
					Attorneys:                 tc.Attorneys,
					ReplacementAttorneys:      tc.ReplacementAttorneys,
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					Attorneys:                           tc.Attorneys,
					ReplacementAttorneys:                tc.ReplacementAttorneys,
					HowAttorneysMakeDecisions:           tc.HowAttorneysMakeDecisions,
					HowShouldReplacementAttorneysStepIn: tc.HowShouldReplacementAttorneysStepIn}).
				Return(nil)

			template := &mockTemplate{}

			err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.ExpectedRedirectUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, template, lpaStore)
		})
	}
}

func TestPostHowShouldReplacementAttorneysStepInFromStore(t *testing.T) {
	testCases := map[string]struct {
		existingWhenStepIn   string
		existingOtherDetails string
		updatedWhenStepIn    string
		updatedOtherDetails  string
		formWhenStepIn       string
		formOtherDetails     string
	}{
		"existing otherDetails not set": {
			existingWhenStepIn:   page.AllCanNoLongerAct,
			existingOtherDetails: "",
			updatedWhenStepIn:    page.SomeOtherWay,
			updatedOtherDetails:  "some details",
			formWhenStepIn:       page.SomeOtherWay,
			formOtherDetails:     "some details",
		},
		"existing otherDetails set": {
			existingWhenStepIn:   page.SomeOtherWay,
			existingOtherDetails: "some details",
			updatedWhenStepIn:    page.OneCanNoLongerAct,
			updatedOtherDetails:  "",
			formWhenStepIn:       page.OneCanNoLongerAct,
			formOtherDetails:     "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"when-to-step-in": {tc.formWhenStepIn},
				"other-details":   {tc.formOtherDetails},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					HowShouldReplacementAttorneysStepIn:        tc.existingWhenStepIn,
					HowShouldReplacementAttorneysStepInDetails: tc.existingOtherDetails,
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					HowShouldReplacementAttorneysStepIn:        tc.updatedWhenStepIn,
					HowShouldReplacementAttorneysStepInDetails: tc.updatedOtherDetails}).
				Return(nil)

			template := &mockTemplate{}

			err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+page.Paths.TaskList, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, template, lpaStore)
		})
	}
}

func TestPostHowShouldReplacementAttorneysStepInFormValidation(t *testing.T) {
	form := url.Values{
		"when-to-step-in": {""},
		"other-details":   {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldReplacementAttorneysStepInData{
			App:    appData,
			Errors: validation.With("when-to-step-in", validation.SelectError{Label: "whenYourReplacementAttorneysStepIn"}),
			Form:   &howShouldReplacementAttorneysStepInForm{},
		}).
		Return(nil)

	err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostHowShouldReplacementAttorneysStepInWhenPutStoreError(t *testing.T) {
	form := url.Values{
		"when-to-step-in": {page.SomeOtherWay},
		"other-details":   {"some details"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			HowShouldReplacementAttorneysStepIn:        "",
			HowShouldReplacementAttorneysStepInDetails: "",
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			HowShouldReplacementAttorneysStepIn:        page.SomeOtherWay,
			HowShouldReplacementAttorneysStepInDetails: "some details"}).
		Return(expectedError)

	template := &mockTemplate{}

	err := HowShouldReplacementAttorneysStepIn(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestHowShouldReplacementAttorneysStepInFormValidate(t *testing.T) {
	testCases := map[string]struct {
		whenToStepIn   string
		otherDetails   string
		expectedErrors validation.List
	}{
		"missing whenToStepIn": {
			whenToStepIn:   "",
			otherDetails:   "",
			expectedErrors: validation.With("when-to-step-in", validation.SelectError{Label: "whenYourReplacementAttorneysStepIn"}),
		},
		"other missing otherDetail": {
			whenToStepIn:   page.SomeOtherWay,
			otherDetails:   "",
			expectedErrors: validation.With("other-details", validation.EnterError{Label: "detailsOfWhenToStepIn"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := howShouldReplacementAttorneysStepInForm{
				WhenToStepIn: tc.whenToStepIn,
				OtherDetails: tc.otherDetails,
			}

			assert.Equal(t, tc.expectedErrors, form.Validate())
		})
	}
}
