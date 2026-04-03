package memberclaim

import (
	"database/sql"
	"errors"
	"strings"

	"familyplan/src/internal/planutil"
	"familyplan/src/internal/support/random"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

const (
	// CollectionName is the PocketBase collection that stores public claim links.
	CollectionName = "member_claim_links"
)

var (
	// ErrClaimLinkNotFound indicates that the claim token is missing or stale.
	ErrClaimLinkNotFound = errors.New("member claim link not found")
	// ErrArtificialMemberUnavailable indicates that the placeholder member can no longer be claimed.
	ErrArtificialMemberUnavailable = errors.New("artificial member is unavailable")
	// ErrAlreadyMember indicates that the claiming user already belongs to the plan.
	ErrAlreadyMember = errors.New("user already belongs to the plan")
)

// Info describes an active artificial-member claim link.
type Info struct {
	Token              string
	PlanRecord         *pbmodels.Record
	PlanID             string
	PlanName           string
	JoinCode           string
	ArtificialMemberID string
	ArtificialName     string
}

// Result describes the destination after a successful claim.
type Result struct {
	JoinCode       string
	PlanName       string
	ArtificialName string
}

// Path returns the canonical public path for a member claim token.
func Path(token string) string {
	return "/claim-member/" + token
}

// ErrorMessage maps expected claim errors to user-facing copy.
func ErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrAlreadyMember):
		return "You're already a member of this plan."
	case errors.Is(err, ErrClaimLinkNotFound), errors.Is(err, ErrArtificialMemberUnavailable):
		return "This claim link is no longer available."
	default:
		return ""
	}
}

// FindByToken loads a claim link by token.
func FindByToken(app *pocketbase.PocketBase, token string) (*pbmodels.Record, error) {
	return FindByTokenWithDao(app.Dao(), token)
}

// FindByTokenWithDao loads a claim link by token using the provided dao.
func FindByTokenWithDao(dao *daos.Dao, token string) (*pbmodels.Record, error) {
	collection, err := dao.FindCollectionByNameOrId(CollectionName)
	if err != nil {
		return nil, err
	}

	record, err := dao.FindFirstRecordByData(collection.Id, "token", token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return record, err
}

// FindForArtificialMember loads the active claim link for an artificial member.
func FindForArtificialMember(app *pocketbase.PocketBase, planID, artificialMemberID string) (*pbmodels.Record, error) {
	return FindForArtificialMemberWithDao(app.Dao(), planID, artificialMemberID)
}

// FindForArtificialMemberWithDao loads the active claim link for an artificial member using the provided dao.
func FindForArtificialMemberWithDao(dao *daos.Dao, planID, artificialMemberID string) (*pbmodels.Record, error) {
	collection, err := dao.FindCollectionByNameOrId(CollectionName)
	if err != nil {
		return nil, err
	}

	filter, err := planutil.BuildEqualsFilter(
		planutil.FilterTerm{Field: "plan_id", Value: planID},
		planutil.FilterTerm{Field: "artificial_member_id", Value: artificialMemberID},
	)
	if err != nil {
		return nil, err
	}

	record, err := dao.FindFirstRecordByFilter(
		collection.Id,
		filter.Expression,
		filter.Params,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return record, err
}

// FindAllForPlan loads all stored claim links for a plan.
func FindAllForPlan(app *pocketbase.PocketBase, planID string) ([]*pbmodels.Record, error) {
	collection, err := app.Dao().FindCollectionByNameOrId(CollectionName)
	if err != nil {
		return nil, err
	}

	filter, err := planutil.BuildEqualsFilter(
		planutil.FilterTerm{Field: "plan_id", Value: planID},
	)
	if err != nil {
		return nil, err
	}

	return app.Dao().FindRecordsByFilter(
		collection.Id,
		filter.Expression,
		"",
		-1,
		0,
		filter.Params,
	)
}

// Ensure makes sure an artificial member has a reusable public claim link.
func Ensure(app *pocketbase.PocketBase, planID, artificialMemberID string) (*pbmodels.Record, error) {
	return EnsureWithDao(app.Dao(), planID, artificialMemberID)
}

// EnsureWithDao makes sure an artificial member has a reusable public claim link using the provided dao.
func EnsureWithDao(dao *daos.Dao, planID, artificialMemberID string) (*pbmodels.Record, error) {
	existing, err := FindForArtificialMemberWithDao(dao, planID, artificialMemberID)
	if err != nil || existing != nil {
		return existing, err
	}

	collection, err := dao.FindCollectionByNameOrId(CollectionName)
	if err != nil {
		return nil, err
	}

	token, err := random.GenerateToken()
	if err != nil {
		return nil, err
	}

	record := pbmodels.NewRecord(collection)
	record.Set("plan_id", planID)
	record.Set("artificial_member_id", artificialMemberID)
	record.Set("token", token)

	if err := dao.SaveRecord(record); err != nil {
		return nil, err
	}

	return record, nil
}

// Lookup resolves a public claim token into the related plan and artificial member metadata.
func Lookup(app *pocketbase.PocketBase, token string) (*Info, error) {
	return LookupWithDao(app.Dao(), token)
}

// LookupWithDao resolves a public claim token into the related plan and artificial member metadata using the provided dao.
func LookupWithDao(dao *daos.Dao, token string) (*Info, error) {
	record, err := FindByTokenWithDao(dao, token)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrClaimLinkNotFound
	}

	plansCollection, err := dao.FindCollectionByNameOrId("family_plans")
	if err != nil {
		return nil, err
	}

	planRecord, err := dao.FindRecordById(plansCollection.Id, record.GetString("plan_id"))
	if errors.Is(err, sql.ErrNoRows) || planRecord == nil {
		return nil, ErrClaimLinkNotFound
	}
	if err != nil {
		return nil, err
	}

	artificialMemberID := record.GetString("artificial_member_id")
	artificialMembership, err := findArtificialMembershipWithDao(dao, planRecord.Id, artificialMemberID)
	if err != nil {
		return nil, err
	}
	if artificialMembership == nil || !artificialMembership.GetDateTime("date_ended").IsZero() {
		return nil, ErrArtificialMemberUnavailable
	}

	return &Info{
		Token:              record.GetString("token"),
		PlanRecord:         planRecord,
		PlanID:             planRecord.Id,
		PlanName:           planRecord.GetString("name"),
		JoinCode:           planRecord.GetString("join_code"),
		ArtificialMemberID: artificialMemberID,
		ArtificialName:     artificialMembership.GetString("name"),
	}, nil
}

// DeleteForArtificialMemberWithDao removes all claim links associated with a placeholder member.
func DeleteForArtificialMemberWithDao(dao *daos.Dao, planID, artificialMemberID string) error {
	collection, err := dao.FindCollectionByNameOrId(CollectionName)
	if err != nil {
		return err
	}

	filter, err := planutil.BuildEqualsFilter(
		planutil.FilterTerm{Field: "plan_id", Value: planID},
		planutil.FilterTerm{Field: "artificial_member_id", Value: artificialMemberID},
	)
	if err != nil {
		return err
	}

	records, err := dao.FindRecordsByFilter(
		collection.Id,
		filter.Expression,
		"",
		-1,
		0,
		filter.Params,
	)
	if err != nil {
		return err
	}

	for _, record := range records {
		if err := dao.DeleteRecord(record); err != nil {
			return err
		}
	}

	return nil
}

// Claim converts an artificial member claim link into a real membership for the provided user.
func Claim(app *pocketbase.PocketBase, token, realUserID string) (Result, error) {
	result := Result{}

	err := app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
		info, err := LookupWithDao(txDao, token)
		if err != nil {
			return err
		}

		if err := TransferArtificialMembership(txDao, info.PlanRecord, info.ArtificialMemberID, realUserID); err != nil {
			return err
		}

		result = Result{
			JoinCode:       info.JoinCode,
			PlanName:       info.PlanName,
			ArtificialName: info.ArtificialName,
		}

		return nil
	})

	return result, err
}

// TransferArtificialMembership replaces a placeholder membership with a real account.
func TransferArtificialMembership(txDao *daos.Dao, planRecord *pbmodels.Record, artificialMemberID, realUserID string) error {
	if planRecord == nil {
		return ErrArtificialMemberUnavailable
	}
	if planutil.IsOwner(planRecord, realUserID) {
		return ErrAlreadyMember
	}

	existingMembership, err := planutil.FindMembershipWithDao(txDao, planRecord.Id, realUserID)
	if err != nil {
		return err
	}
	if existingMembership != nil {
		return ErrAlreadyMember
	}

	usersCollection, err := txDao.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	realUserRecord, err := txDao.FindRecordById(usersCollection.Id, realUserID)
	if errors.Is(err, sql.ErrNoRows) || realUserRecord == nil {
		return ErrClaimLinkNotFound
	}
	if err != nil {
		return err
	}

	artificialMembership, err := findArtificialMembershipWithDao(txDao, planRecord.Id, artificialMemberID)
	if err != nil {
		return err
	}
	if artificialMembership == nil || !artificialMembership.GetDateTime("date_ended").IsZero() {
		return ErrArtificialMemberUnavailable
	}

	paymentsCollection, err := txDao.FindCollectionByNameOrId("payments")
	if err != nil {
		return err
	}

	artificialPaymentsFilter, err := planutil.BuildEqualsFilter(
		planutil.FilterTerm{Field: "plan_id", Value: planRecord.Id},
		planutil.FilterTerm{Field: "user_id", Value: artificialMemberID},
	)
	if err != nil {
		return err
	}

	artificialPayments, err := txDao.FindRecordsByFilter(
		paymentsCollection.Id,
		artificialPaymentsFilter.Expression,
		"",
		-1,
		0,
		artificialPaymentsFilter.Params,
	)
	if err != nil {
		return err
	}

	for _, payment := range artificialPayments {
		payment.Set("user_id", realUserID)
		if err := txDao.SaveRecord(payment); err != nil {
			return err
		}
	}

	if strings.TrimSpace(realUserRecord.GetString("name")) == "" {
		if artificialName := strings.TrimSpace(artificialMembership.GetString("name")); artificialName != "" {
			realUserRecord.Set("name", artificialName)
			if err := txDao.SaveRecord(realUserRecord); err != nil {
				return err
			}
		}
	}

	if err := txDao.DeleteRecord(artificialMembership); err != nil {
		return err
	}

	membershipsCollection, err := txDao.FindCollectionByNameOrId("memberships")
	if err != nil {
		return err
	}

	newMembership := pbmodels.NewRecord(membershipsCollection)
	newMembership.Set("plan_id", planRecord.Id)
	newMembership.Set("user_id", realUserID)
	newMembership.Set("is_artificial", false)
	newMembership.Set("created", artificialMembership.GetDateTime("created"))
	if err := txDao.SaveRecord(newMembership); err != nil {
		return err
	}

	request, err := planutil.FindJoinRequestWithDao(txDao, planRecord.Id, realUserID)
	if err != nil {
		return err
	}
	if request != nil {
		if err := txDao.DeleteRecord(request); err != nil {
			return err
		}
	}

	return DeleteForArtificialMemberWithDao(txDao, planRecord.Id, artificialMemberID)
}

func findArtificialMembershipWithDao(dao *daos.Dao, planID, artificialMemberID string) (*pbmodels.Record, error) {
	membershipsCollection, err := dao.FindCollectionByNameOrId("memberships")
	if err != nil {
		return nil, err
	}

	filter, err := planutil.BuildEqualsFilter(
		planutil.FilterTerm{Field: "plan_id", Value: planID},
		planutil.FilterTerm{Field: "user_id", Value: artificialMemberID},
		planutil.FilterTerm{Field: "is_artificial", Value: true},
	)
	if err != nil {
		return nil, err
	}

	record, err := dao.FindFirstRecordByFilter(
		membershipsCollection.Id,
		filter.Expression,
		filter.Params,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return record, err
}
