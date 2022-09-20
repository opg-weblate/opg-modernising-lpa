package page

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"

	"github.com/ministryofjustice/opg-go-common/template"
)

type aboutPaymentData struct {
	App    AppData
	Errors map[string]string
}

func AboutPayment(logger Logger, tmpl template.Template, sessionStore sessions.Store, payClient pay.PayClient, appPublicUrl string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		data := &aboutPaymentData{
			App: appData,
		}

		if r.Method == http.MethodPost {
			//createPaymentBody := pay.CreatePaymentBody{
			//	Amount:      0,
			//	Reference:   "abc",
			//	Description: "A payment",
			//	ReturnUrl:   appPublicUrl + "/payment-confirmation",
			//	Email:       "a@b.com",
			//	Language:    "en",
			//}

			//resp, err := payClient.CreatePayment(createPaymentBody)

			//if err != nil {
			//	logger.Print(fmt.Sprintf("Error creating payment: %s", err.Error()))
			//	return err
			//}

			//nextUrl := resp.Links["next_url"].Href

			nextUrl := "http://localhost:4010"
			secureCookies := strings.HasPrefix(nextUrl, "https:")

			cookieOptions := &sessions.Options{
				// Should we lock this down to payment confirmation page?
				Path: "/",
				// A payment can be resumed up to 90 minutes after creation
				MaxAge:   int(time.Minute * 90 / time.Second),
				SameSite: http.SameSiteLaxMode,
				HttpOnly: true,
				Secure:   secureCookies,
			}

			session := sessions.NewSession(sessionStore, "pay")
			session.Values = map[interface{}]interface{}{
				//"paymentId": resp.PaymentId,
				"paymentId": "abc123",
			}
			session.Options = cookieOptions

			//session.Values = map[interface{}]interface{}{"paymentId": resp.PaymentId}
			session.Values = map[interface{}]interface{}{"paymentId": "abc123"}
			if err := sessionStore.Save(r, w, session); err != nil {
				return err
			}

			// If URL matches expected domain for GOV UK PAY redirect there. If not, redirect to the confirmation code and carry on with flow.
			if strings.Contains(nextUrl, "https://publicapi.payments.service.gov.uk") {
				http.Redirect(w, r, nextUrl, http.StatusFound)
			} else {
				http.Redirect(w, r, "/payment-confirmation", http.StatusFound)
			}

		}

		return tmpl(w, data)
	}
}
