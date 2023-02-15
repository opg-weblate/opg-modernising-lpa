package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"

	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCertificateProviderYourDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		ID: "lpa-id",
	}

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &yourDetailsData{
			App:  TestAppData,
			Lpa:  lpa,
			Form: &yourDetailsForm{},
		}).
		Return(nil)

	err := YourDetails(template.Func, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetCertificateProviderYourDetailsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		ID: "lpa-id",
		CertificateProviderProvidedDetails: actor.CertificateProvider{
			Email:       "a@example.org",
			Mobile:      "07535111222",
			DateOfBirth: date.New("1997", "1", "2"),
		},
	}
	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &yourDetailsData{
			App: TestAppData,
			Lpa: lpa,
			Form: &yourDetailsForm{
				Email:  "a@example.org",
				Mobile: "07535111222",
				Dob:    date.New("1997", "1", "2"),
			},
		}).
		Return(nil)

	err := YourDetails(template.Func, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetCertificateProviderYourDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, ExpectedError)

	err := YourDetails(nil, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetCertificateProviderYourDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		ID: "lpa-id",
	}

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &yourDetailsData{
			App:  TestAppData,
			Lpa:  lpa,
			Form: &yourDetailsForm{},
		}).
		Return(ExpectedError)

	err := YourDetails(template.Func, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostCertificateProviderYourDetails(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form url.Values
		cp   actor.CertificateProvider
	}{
		"valid": {
			form: url.Values{
				"email":               {"name@example.com"},
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			cp: actor.CertificateProvider{
				DateOfBirth: date.New(validBirthYear, "1", "2"),
				Email:       "name@example.com",
				Mobile:      "07535111222",
			},
		},
		"warning ignored": {
			form: url.Values{
				"email":               {"name@example.com"},
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
			},
			cp: actor.CertificateProvider{
				DateOfBirth: date.New("1900", "1", "2"),
				Email:       "name@example.com",
				Mobile:      "07535111222",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := &MockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{ID: "lpa-id"}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					ID:                                 "lpa-id",
					CertificateProviderProvidedDetails: tc.cp,
				}).
				Return(nil)

			sessionStore := &MockSessionsStore{}
			sessionStore.
				On("Get", r, "session").
				Return(&sessions.Session{Values: map[any]any{"certificate-provider": &sesh.CertificateProviderSession{Sub: "xyz", LpaID: "lpa-id"}}}, nil)

			err := YourDetails(nil, lpaStore)(TestAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.CertificateProviderYourAddress, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostCertificateProviderYourDetailsWhenInputRequired(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *yourDetailsData) bool
	}{
		"validation error": {
			form: url.Values{
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			dataMatcher: func(t *testing.T, data *yourDetailsData) bool {
				return assert.Equal(t, validation.With("email", validation.EnterError{Label: "email"}), data.Errors)
			},
		},
		"dob warning": {
			form: url.Values{
				"email":               {"name@example.com"},
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
			},
			dataMatcher: func(t *testing.T, data *yourDetailsData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
			},
		},
		"dob warning ignored but other errors": {
			form: url.Values{
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsOver100"},
			},
			dataMatcher: func(t *testing.T, data *yourDetailsData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning) &&
					assert.Equal(t, validation.With("email", validation.EnterError{Label: "email"}), data.Errors)
			},
		},
		"other dob warning ignored": {
			form: url.Values{
				"email":               {"name@example.com"},
				"mobile":              {"07535111222"},
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
				"ignore-dob-warning":  {"dateOfBirthIsUnder18"},
			},
			dataMatcher: func(t *testing.T, data *yourDetailsData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := &MockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{ID: "lpa-id"}, nil)

			template := &MockTemplate{}
			template.
				On("Func", w, mock.MatchedBy(func(data *yourDetailsData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			sessionStore := &MockSessionsStore{}
			sessionStore.
				On("Get", r, "session").
				Return(&sessions.Session{Values: map[any]any{"certificate-provider": &sesh.CertificateProviderSession{Sub: "xyz", LpaID: "lpa-id"}}}, nil)

			err := YourDetails(template.Func, lpaStore)(TestAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostYourDetailsWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"email":               {"name@example.com"},
		"mobile":              {"07535111222"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1999"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, ExpectedError)

	err := YourDetails(nil, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestReadYourDetailsForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"email":               {"name@example.com"},
		"mobile":              {"07535111222"},
		"date-of-birth-day":   {"2"},
		"date-of-birth-month": {"1"},
		"date-of-birth-year":  {"1990"},
		"ignore-dob-warning":  {"xyz"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readYourDetailsForm(r)

	assert.Equal("name@example.com", result.Email)
	assert.Equal("07535111222", result.Mobile)
	assert.Equal(date.New("1990", "1", "2"), result.Dob)
	assert.Equal("xyz", result.IgnoreDobWarning)
}

func TestYourDetailsFormValidate(t *testing.T) {
	now := date.Today()
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		form   *yourDetailsForm
		errors validation.List
	}{
		"valid": {
			form: &yourDetailsForm{
				Dob:              validDob,
				Mobile:           "07535999222",
				Email:            "name@example.org",
				IgnoreDobWarning: "xyz",
			},
		},
		"missing-all": {
			form: &yourDetailsForm{},
			errors: validation.
				With("date-of-birth", validation.EnterError{Label: "dateOfBirth"}).
				With("mobile", validation.EnterError{Label: "mobile"}).
				With("email", validation.EnterError{Label: "email"}),
		},
		"future-dob": {
			form: &yourDetailsForm{
				Mobile: "07535999222",
				Email:  "name@example.org",
				Dob:    now.AddDate(0, 0, 1),
			},
			errors: validation.With("date-of-birth", validation.DateMustBePastError{Label: "dateOfBirth"}),
		},
		"invalid-dob": {
			form: &yourDetailsForm{
				Mobile: "07535999222",
				Email:  "name@example.org",
				Dob:    date.New("2000", "22", "2"),
			},
			errors: validation.With("date-of-birth", validation.DateMustBeRealError{Label: "dateOfBirth"}),
		},
		"invalid-missing-dob": {
			form: &yourDetailsForm{
				Mobile: "07535999222",
				Email:  "name@example.org",
				Dob:    date.New("1", "", "1"),
			},
			errors: validation.With("date-of-birth", validation.DateMissingError{Label: "dateOfBirth", MissingMonth: true}),
		},
		"invalid-mobile-format": {
			form: &yourDetailsForm{
				Mobile: "123",
				Email:  "name@example.org",
				Dob:    validDob,
			},
			errors: validation.With("mobile", validation.MobileError{Label: "mobile"}),
		},
		"invalid-email-format": {
			form: &yourDetailsForm{
				Mobile: "07535999222",
				Email:  "name@",
				Dob:    validDob,
			},
			errors: validation.With("email", validation.EmailError{Label: "email"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
