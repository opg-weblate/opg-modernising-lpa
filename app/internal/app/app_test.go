package app

import (
	"log"
	"net/http"
	"testing"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	app := App(&log.Logger{}, localize.Localizer{}, localize.En, template.Templates{}, nil, nil, "http://public.url", &pay.Client{}, &identity.YotiClient{}, "yoti-scenario-id", &notify.Client{}, &place.Client{}, page.RumConfig{}, "?%3fNEI0t9MN", page.Paths, &onelogin.Client{})

	assert.Implements(t, (*http.Handler)(nil), app)
}
