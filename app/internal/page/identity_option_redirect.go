package page

import (
	"net/http"
)

func IdentityOptionRedirect(dataStore DataStore) Handler {
	identityOptionPaths := map[IdentityOption]string{
		Yoti: identityWithYotiPath,
		// obviously the below will change eventually
		Passport:                 identityWithYotiPath,
		DrivingLicence:           identityWithYotiPath,
		GovernmentGatewayAccount: identityWithYotiPath,
		DwpAccount:               identityWithYotiPath,
		OnlineBankAccount:        identityWithYotiPath,
		UtilityBill:              identityWithYotiPath,
		CouncilTaxBill:           identityWithYotiPath,
	}

	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
			return err
		}

		switch lpa.IdentityOptions.Current {
		case 0:
			appData.Lang.Redirect(w, r, identityOptionPaths[lpa.IdentityOptions.First], http.StatusFound)
			lpa.IdentityOptions.Current++

			if err := dataStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
				return err
			}
		case 1:
			appData.Lang.Redirect(w, r, identityOptionPaths[lpa.IdentityOptions.Second], http.StatusFound)
			lpa.IdentityOptions.Current++

			if err := dataStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
				return err
			}
		default:
			appData.Lang.Redirect(w, r, whatHappensWhenSigningPath, http.StatusFound)
		}

		return nil
	}
}