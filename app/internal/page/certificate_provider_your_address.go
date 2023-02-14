package page

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type cpYourAddressData struct {
	App       AppData
	Errors    validation.List
	Addresses []place.Address
	Form      *addressForm
}

func CpYourAddress(logger Logger, tmpl template.Template, addressClient AddressClient, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &cpYourAddressData{
			App:  appData,
			Form: &addressForm{},
		}

		if lpa.You.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &lpa.You.Address
		}

		if r.Method == http.MethodPost {
			data.Form = readAddressForm(r)
			data.Errors = data.Form.Validate()

			if data.Form.Action == "manual" && data.Errors.None() {
				lpa.You.Address = *data.Form.Address
				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, Paths.WhoIsTheLpaFor)
			}

			if data.Form.Action == "select" && data.Errors.None() {
				data.Form.Action = "manual"
			}

			if data.Form.Action == "lookup" && data.Errors.None() ||
				data.Form.Action == "select" && data.Errors.Any() {
				addresses, err := addressClient.LookupPostcode(r.Context(), data.Form.LookupPostcode)
				if err != nil {
					logger.Print(err)

					if errors.As(err, &place.BadRequestError{}) {
						data.Errors.Add("lookup-postcode", validation.EnterError{Label: "invalidPostcode"})
					} else {
						data.Errors.Add("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"})
					}
				} else if len(addresses) == 0 {
					data.Errors.Add("lookup-postcode", validation.CustomError{Label: "noAddressesFound"})
				}

				data.Addresses = addresses
			}
		}

		if r.Method == http.MethodGet {
			action := r.FormValue("action")
			if action == "manual" {
				data.Form.Action = "manual"
				data.Form.Address = &place.Address{}
			}
		}

		return tmpl(w, data)
	}
}