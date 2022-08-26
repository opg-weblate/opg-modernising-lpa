package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type donorAddressData struct {
	Page      string
	L         localize.Localizer
	Lang      Lang
	Errors    map[string]string
	Addresses []Address
	Form      *donorAddressForm
}

func DonorAddress(logger Logger, localizer localize.Localizer, lang Lang, tmpl template.Template, addressClient AddressClient, dataStore DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := &donorAddressData{
			Page: donorAddressPath,
			L:    localizer,
			Lang: lang,
			Form: &donorAddressForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readDonorAddressForm(r)
			data.Errors = data.Form.Validate()

			if (data.Form.Action == "manual" || data.Form.Action == "select") && len(data.Errors) == 0 {
				dataStore.Save(data.Form.Address)
				lang.Redirect(w, r, whoIsTheLpaForPath, http.StatusFound)
				return
			}

			if data.Form.Action == "lookup" && len(data.Errors) == 0 ||
				data.Form.Action == "select" && len(data.Errors) > 0 {
				addresses, err := addressClient.LookupPostcode(data.Form.LookupPostcode)
				if err != nil {
					logger.Print(err)
					data.Errors["lookup-postcode"] = "couldNotLookupPostcode"
				}
				data.Addresses = addresses
			}
		}

		if r.Method == http.MethodGet {
			action := r.FormValue("action")
			if action == "manual" {
				data.Form.Action = "manual"
				data.Form.Address = &Address{}
			}
		}

		if err := tmpl(w, data); err != nil {
			logger.Print(err)
		}
	}
}

type donorAddressForm struct {
	Action         string
	LookupPostcode string
	Address        *Address
}

func readDonorAddressForm(r *http.Request) *donorAddressForm {
	d := &donorAddressForm{}
	d.Action = r.PostFormValue("action")

	switch d.Action {
	case "lookup":
		d.LookupPostcode = postFormString(r, "lookup-postcode")

	case "select":
		d.LookupPostcode = postFormString(r, "lookup-postcode")
		selectAddress := r.PostFormValue("select-address")
		if selectAddress != "" {
			d.Address = DecodeAddress(selectAddress)
		}

	case "manual":
		d.Address = &Address{
			Line1:      postFormString(r, "address-line-1"),
			Line2:      postFormString(r, "address-line-2"),
			TownOrCity: postFormString(r, "address-town"),
			Postcode:   postFormString(r, "address-postcode"),
		}
	}

	return d
}

func (d *donorAddressForm) Validate() map[string]string {
	errors := map[string]string{}

	switch d.Action {
	case "lookup":
		if d.LookupPostcode == "" {
			errors["lookup-postcode"] = "enterPostcode"
		}

	case "select":
		if d.Address == nil {
			errors["select-address"] = "selectYourAddress"
		}

	case "manual":
		if d.Address.Line1 == "" {
			errors["address-line-1"] = "enterYourAddress"
		}
		if d.Address.TownOrCity == "" {
			errors["address-town"] = "enterYourTownOrCity"
		}
		if d.Address.Postcode == "" {
			errors["address-postcode"] = "enterYourPostcode"
		}
	}

	return errors
}