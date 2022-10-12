package page

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/ordnance_survey"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
)

const (
	PayCookieName              = "pay"
	PayCookiePaymentIdValueKey = "paymentId"
	CostOfLpaPence             = 8200
)

type TaskState int

const (
	TaskNotStarted TaskState = iota
	TaskInProgress
	TaskCompleted
)

type Lpa struct {
	ID                       string
	You                      Person
	Attorney                 Attorney
	CertificateProvider      CertificateProvider
	WhoFor                   string
	Contact                  []string
	Type                     string
	WantReplacementAttorneys string
	WhenCanTheLpaBeUsed      string
	Restrictions             string
	Tasks                    Tasks
	Checked                  bool
	HappyToShare             bool
	PaymentDetails           PaymentDetails
	CheckedAgain             bool
	ConfirmFreeWill          bool
	SignatureCode            string
	EnteredSignatureCode     string
	SignatureEmailID         string
	IdentityOptions          IdentityOptions
	YotiUserData             identity.UserData
}

type PaymentDetails struct {
	PaymentReference string
	PaymentId        string
}

type Tasks struct {
	WhenCanTheLpaBeUsed        TaskState
	Restrictions               TaskState
	CertificateProvider        TaskState
	CheckYourLpa               TaskState
	PayForLpa                  TaskState
	ConfirmYourIdentityAndSign TaskState
}

type Person struct {
	FirstNames  string
	LastName    string
	Email       string
	OtherNames  string
	DateOfBirth time.Time
	Address     Address
}

type Attorney struct {
	FirstNames  string
	LastName    string
	Email       string
	DateOfBirth time.Time
	Address     Address
}

type CertificateProvider struct {
	FirstNames              string
	LastName                string
	Email                   string
	DateOfBirth             time.Time
	Relationship            []string
	RelationshipDescription string
	RelationshipLength      string
}

type Address struct {
	Line1      string
	Line2      string
	Line3      string
	TownOrCity string
	Postcode   string
}

type AddressClient interface {
	LookupPostcode(ctx context.Context, postcode string) (ordnance_survey.PostcodeLookupResponse, error)
}

func (a Address) Encode() string {
	x, _ := json.Marshal(a)
	return string(x)
}

func DecodeAddress(s string) *Address {
	var v Address
	json.Unmarshal([]byte(s), &v)
	return &v
}

func (a Address) String() string {
	var parts []string

	if a.Line1 != "" {
		parts = append(parts, a.Line1)
	}
	if a.Line2 != "" {
		parts = append(parts, a.Line2)
	}
	if a.Line3 != "" {
		parts = append(parts, a.Line3)
	}
	if a.TownOrCity != "" {
		parts = append(parts, a.TownOrCity)
	}
	if a.Postcode != "" {
		parts = append(parts, a.Postcode)
	}

	return strings.Join(parts, ", ")
}

func TransformAddressDetailsToAddress(ad ordnance_survey.AddressDetails) Address {
	a := Address{}

	if len(ad.BuildingName) > 0 {
		a.Line1 = ad.BuildingName

		if len(ad.BuildingNumber) > 0 {
			a.Line2 = fmt.Sprintf("%s %s", ad.BuildingNumber, ad.ThoroughFareName)
		} else {
			a.Line2 = ad.ThoroughFareName
		}

		a.Line3 = ad.DependentLocality
	} else {
		a.Line1 = fmt.Sprintf("%s %s", ad.BuildingNumber, ad.ThoroughFareName)
		a.Line2 = ad.DependentLocality
	}

	a.TownOrCity = ad.Town
	a.Postcode = ad.Postcode

	return a
}

func TransformAddressDetailsToAddresses(ads []ordnance_survey.AddressDetails) []Address {
	var addresses []Address

	for _, ad := range ads {
		addresses = append(addresses, TransformAddressDetailsToAddress(ad))
	}

	return addresses
}

type Date struct {
	Day   string
	Month string
	Year  string
}

func readDate(t time.Time) Date {
	return Date{
		Day:   t.Format("2"),
		Month: t.Format("1"),
		Year:  t.Format("2006"),
	}
}

type LpaStore interface {
	Get(context.Context, string) (Lpa, error)
	Put(context.Context, string, Lpa) error
}

type lpaStore struct {
	dataStore DataStore
	randomInt func(int) int
}

func (s *lpaStore) Get(ctx context.Context, sessionID string) (Lpa, error) {
	var lpa Lpa
	if err := s.dataStore.Get(ctx, sessionID, &lpa); err != nil {
		return lpa, err
	}

	if lpa.ID == "" {
		lpa.ID = "10" + strconv.Itoa(s.randomInt(100000))
	}

	// we don't ask for this yet but it is needed for emailing a signature code
	if lpa.You.Email == "" {
		lpa.You.Email = "simulate-delivered@notifications.service.gov.uk"
	}

	return lpa, nil
}

func (s *lpaStore) Put(ctx context.Context, sessionID string, lpa Lpa) error {
	return s.dataStore.Put(ctx, sessionID, lpa)
}
