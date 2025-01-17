package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChoosePeopleToNotifyAddress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			Form:           &addressForm{},
			PersonToNotify: personToNotify,
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChoosePeopleToNotifyAddressWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetChoosePeopleToNotifyAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: address,
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &addressForm{
				Action:  "manual",
				Address: &address,
			},
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChoosePeopleToNotifyAddressManual(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?action=manual&id=123", nil)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: address,
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App: appData,
			Form: &addressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			PersonToNotify: personToNotify,
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetChoosePeopleToNotifyAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			Form:           &addressForm{},
			PersonToNotify: personToNotify,
		}).
		Return(expectedError)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostChoosePeopleToNotifyAddressManual(t *testing.T) {
	form := url.Values{
		"action":           {"manual"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-line-3":   {"c"},
		"address-town":     {"d"},
		"address-postcode": {"e"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
			Tasks:          page.Tasks{PeopleToNotify: page.TaskInProgress},
		}, nil)

	personToNotify.Address = address

	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
			Tasks:          page.Tasks{PeopleToNotify: page.TaskCompleted},
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyAddressManualWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"action":           {"manual"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-line-3":   {"c"},
		"address-town":     {"d"},
		"address-postcode": {"e"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	personToNotify.Address = address

	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
			Tasks:          page.Tasks{PeopleToNotify: page.TaskCompleted},
		}).
		Return(expectedError)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyAddressManualFromStore(t *testing.T) {
	form := url.Values{
		"action":           {"manual"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-line-3":   {"c"},
		"address-town":     {"d"},
		"address-postcode": {"e"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	personToNotify := actor.PersonToNotify{
		ID:         "123",
		FirstNames: "John",
		Address:    place.Address{Line1: "line1"},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
			Tasks:          page.Tasks{PeopleToNotify: page.TaskInProgress},
		}, nil)

	personToNotify.Address = address

	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
			Tasks:          page.Tasks{PeopleToNotify: page.TaskCompleted},
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyAddressSelect(t *testing.T) {
	form := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
		"select-address":  {address.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	personToNotify := actor.PersonToNotify{
		ID:         "123",
		FirstNames: "John",
		Address:    place.Address{Line1: "abc"},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	updatedPersonToNotify := actor.PersonToNotify{
		ID: "123",
		Address: place.Address{
			Line1:      "a",
			Line2:      "b",
			Line3:      "c",
			TownOrCity: "d",
			Postcode:   "e",
		},
		FirstNames: "John",
	}

	lpaStore.
		On("Put", r.Context(), &page.Lpa{PeopleToNotify: actor.PeopleToNotify{updatedPersonToNotify}}).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &addressForm{
				Action:         "manual",
				LookupPostcode: "NG1",
				Address:        &address,
			},
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostChoosePeopleToNotifyAddressSelectWhenValidationError(t *testing.T) {
	form := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	addresses := []place.Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &addressForm{
				Action:         "select",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
			Errors:    validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostChoosePeopleToNotifyAddressLookup(t *testing.T) {
	form := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	addresses := []place.Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	personToNotify := actor.PersonToNotify{
		ID:         "123",
		Address:    place.Address{},
		FirstNames: "John",
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &addressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template)
}

func TestPostChoosePeopleToNotifyAddressLookupError(t *testing.T) {
	form := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return([]place.Address{}, expectedError)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &addressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(logger, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template, logger)
}

func TestPostChoosePeopleToNotifyAddressInvalidPostcodeError(t *testing.T) {
	w := httptest.NewRecorder()
	invalidPostcodeErr := place.BadRequestError{
		Statuscode: 400,
		Message:    "invalid postcode",
	}

	form := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	logger := &mockLogger{}
	logger.
		On("Print", invalidPostcodeErr)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "XYZ").
		Return([]place.Address{}, invalidPostcodeErr)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &addressForm{
				Action:         "lookup",
				LookupPostcode: "XYZ",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.EnterError{Label: "invalidPostcode"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(logger, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template, logger)
}

func TestPostChoosePeopleToNotifyAddressPostcodeNoAddresses(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	logger := &mockLogger{}

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "XYZ").
		Return([]place.Address{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &addressForm{
				Action:         "lookup",
				LookupPostcode: "XYZ",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "noAddressesFound"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(logger, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template, logger)
}

func TestPostChoosePeopleToNotifyAddressLookupWhenValidationError(t *testing.T) {
	form := url.Values{
		"action": {"lookup"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &addressForm{
				Action: "lookup",
			},
			Errors: validation.With("lookup-postcode", validation.EnterError{Label: "aPostcode"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostPersonToNotifyAddressManuallyFromAnotherPage(t *testing.T) {
	testcases := map[string]struct {
		requestUrl      string
		expectedNextUrl string
	}{
		"with from value": {
			"/?from=/test&id=123",
			"/lpa/lpa-id/test",
		},
		"without from value": {
			"/?from=&id=123",
			"/lpa/lpa-id" + page.Paths.ChoosePeopleToNotifySummary,
		},
		"missing from key": {
			"/?id=123",
			"/lpa/lpa-id" + page.Paths.ChoosePeopleToNotifySummary,
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			form := url.Values{
				"action":           {"manual"},
				"address-line-1":   {"a"},
				"address-town":     {"b"},
				"address-postcode": {"c"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, tc.requestUrl, strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					PeopleToNotify: actor.PeopleToNotify{
						{
							ID: "123",
							Address: place.Address{
								Line1:      "a",
								TownOrCity: "b",
								Postcode:   "c",
							},
						},
					},
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					PeopleToNotify: actor.PeopleToNotify{
						{
							ID: "123",
							Address: place.Address{
								Line1:      "a",
								TownOrCity: "b",
								Postcode:   "c",
							},
						},
					},
					Tasks: page.Tasks{PeopleToNotify: page.TaskCompleted},
				}).
				Return(nil)

			err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedNextUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}
