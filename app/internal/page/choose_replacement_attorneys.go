package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseReplacementAttorneysData struct {
	App         AppData
	Errors      validation.List
	Form        *chooseAttorneysForm
	DobWarning  string
	NameWarning *sameActorNameWarning
}

func ChooseReplacementAttorneys(tmpl template.Template, lpaStore LpaStore, randomString func(int) string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		addAnother := r.FormValue("addAnother") == "1"
		attorney, attorneyFound := lpa.GetReplacementAttorney(r.URL.Query().Get("id"))

		if r.Method == http.MethodGet && len(lpa.ReplacementAttorneys) > 0 && attorneyFound == false && addAnother == false {
			return appData.Redirect(w, r, lpa, Paths.ChooseReplacementAttorneysSummary)
		}

		data := &chooseReplacementAttorneysData{
			App: appData,
			Form: &chooseAttorneysForm{
				FirstNames: attorney.FirstNames,
				LastName:   attorney.LastName,
				Email:      attorney.Email,
				Dob:        attorney.DateOfBirth,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			var nameWarning *sameActorNameWarning
			if matchingActor := replacementAttorneyMatches(lpa, attorney.ID, data.Form.FirstNames, data.Form.LastName); matchingActor != "" {
				nameWarning = &sameActorNameWarning{
					Key:        matchingActor,
					Type:       "aReplacementAttorney",
					FirstNames: data.Form.FirstNames,
					LastName:   data.Form.LastName,
				}
			}

			if data.Errors.Any() || data.Form.IgnoreDobWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if data.Errors.Any() || data.Form.IgnoreNameWarning != nameWarning.String() {
				data.NameWarning = nameWarning
			}

			if data.Errors.None() && data.DobWarning == "" && data.NameWarning == nil {
				if attorneyFound == false {
					attorney = Attorney{
						FirstNames:  data.Form.FirstNames,
						LastName:    data.Form.LastName,
						Email:       data.Form.Email,
						DateOfBirth: data.Form.Dob,
						ID:          randomString(8),
					}

					lpa.ReplacementAttorneys = append(lpa.ReplacementAttorneys, attorney)
				} else {
					attorney.FirstNames = data.Form.FirstNames
					attorney.LastName = data.Form.LastName
					attorney.Email = data.Form.Email
					attorney.DateOfBirth = data.Form.Dob

					lpa.PutReplacementAttorney(attorney)
				}

				if !attorneyFound {
					lpa.Tasks.ChooseReplacementAttorneys = TaskInProgress
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				from := r.FormValue("from")

				if from == "" {
					from = fmt.Sprintf("%s?id=%s", appData.Paths.ChooseReplacementAttorneysAddress, attorney.ID)
				}

				return appData.Redirect(w, r, lpa, from)
			}
		}

		return tmpl(w, data)
	}
}

func replacementAttorneyMatches(lpa *Lpa, id, firstNames, lastName string) string {
	if lpa.You.FirstNames == firstNames && lpa.You.LastName == lastName {
		return "errorDonorMatchesActor"
	}

	for _, attorney := range lpa.Attorneys {
		if attorney.FirstNames == firstNames && attorney.LastName == lastName {
			return "errorAttorneyMatchesActor"
		}
	}

	for _, attorney := range lpa.ReplacementAttorneys {
		if attorney.ID != id && attorney.FirstNames == firstNames && attorney.LastName == lastName {
			return "errorReplacementAttorneyMatchesReplacementAttorney"
		}
	}

	if lpa.CertificateProvider.FirstNames == firstNames && lpa.CertificateProvider.LastName == lastName {
		return "errorCertificateProviderMatchesActor"
	}

	for _, person := range lpa.PeopleToNotify {
		if person.FirstNames == firstNames && person.LastName == lastName {
			return "errorPersonToNotifyMatchesActor"
		}
	}

	return ""
}
