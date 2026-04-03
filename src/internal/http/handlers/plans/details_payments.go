package plans

import (
	"familyplan/src/internal/domain"
	"familyplan/src/internal/planutil"

	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

func loadPendingPayments(app *pocketbase.PocketBase, planID string) ([]domain.Payment, error) {
	return loadPaymentsByTerms(app, planID, -1, 0,
		planutil.FilterTerm{Field: "plan_id", Value: planID},
		planutil.FilterTerm{Field: "status", Value: "pending"},
	)
}

func loadAllPaymentsPage(app *pocketbase.PocketBase, planID string, page, pageSize int) ([]domain.Payment, domain.MemberPaymentsPagination, error) {
	payments, err := loadPaymentsByTerms(app, planID, pageSize+1, (page-1)*pageSize,
		planutil.FilterTerm{Field: "plan_id", Value: planID},
	)
	if err != nil {
		return nil, domain.MemberPaymentsPagination{}, err
	}

	hasNext := len(payments) > pageSize
	if hasNext {
		payments = payments[:pageSize]
	}

	return payments, buildMemberPaymentsPagination(page, hasNext), nil
}

func loadUserPayments(app *pocketbase.PocketBase, planID, userID string) ([]domain.Payment, error) {
	paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
	if err != nil {
		return nil, err
	}

	filter, err := planutil.BuildEqualsFilter(
		planutil.FilterTerm{Field: "plan_id", Value: planID},
		planutil.FilterTerm{Field: "user_id", Value: userID},
	)
	if err != nil {
		return nil, err
	}

	paymentRecords, err := app.Dao().FindRecordsByFilter(
		paymentsCollection.Id,
		filter.Expression,
		"-created",
		20,
		0,
		filter.Params,
	)
	if err != nil {
		return nil, err
	}

	payments := make([]domain.Payment, 0, len(paymentRecords))
	for _, paymentRecord := range paymentRecords {
		payments = append(payments, buildPayment(paymentRecord, "", ""))
	}

	return payments, nil
}

func loadPaymentsByTerms(app *pocketbase.PocketBase, planID string, limit, offset int, terms ...planutil.FilterTerm) ([]domain.Payment, error) {
	paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
	if err != nil {
		return nil, err
	}

	filter, err := planutil.BuildEqualsFilter(terms...)
	if err != nil {
		return nil, err
	}

	paymentRecords, err := app.Dao().FindRecordsByFilter(
		paymentsCollection.Id,
		filter.Expression,
		"-created",
		limit,
		offset,
		filter.Params,
	)
	if err != nil {
		return nil, err
	}

	identities, err := loadPaymentIdentities(app, planID, paymentUserIDs(paymentRecords))
	if err != nil {
		return nil, err
	}

	payments := make([]domain.Payment, 0, len(paymentRecords))
	for _, paymentRecord := range paymentRecords {
		identity, ok := identities[paymentRecord.GetString("user_id")]
		if !ok {
			continue
		}

		payments = append(payments, buildPayment(paymentRecord, identity.Username, identity.Name))
	}

	return payments, nil
}

type paymentIdentity struct {
	Username string
	Name     string
}

func loadPaymentIdentities(app *pocketbase.PocketBase, planID string, userIDs []string) (map[string]paymentIdentity, error) {
	identities := make(map[string]paymentIdentity, len(userIDs))
	remainingUserIDs := uniqueNonEmptyStrings(userIDs)
	if len(remainingUserIDs) == 0 {
		return identities, nil
	}

	membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
	if err != nil {
		return nil, err
	}

	filter, err := planutil.BuildEqualsFilter(
		planutil.FilterTerm{Field: "plan_id", Value: planID},
		planutil.FilterTerm{Field: "is_artificial", Value: true},
	)
	if err != nil {
		return nil, err
	}

	artificialMemberships, err := app.Dao().FindRecordsByFilter(
		membershipsCollection.Id,
		filter.Expression,
		"",
		-1,
		0,
		filter.Params,
	)
	if err != nil {
		return nil, err
	}

	unresolved := make(map[string]struct{}, len(remainingUserIDs))
	for _, userID := range remainingUserIDs {
		unresolved[userID] = struct{}{}
	}

	for _, artificialMembership := range artificialMemberships {
		userID := artificialMembership.GetString("user_id")
		if _, ok := unresolved[userID]; !ok {
			continue
		}

		identities[userID] = paymentIdentity{
			Name: artificialMembership.GetString("name"),
		}
		delete(unresolved, userID)
	}

	if len(unresolved) == 0 {
		return identities, nil
	}

	usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		return nil, err
	}

	userRecords, err := app.Dao().FindRecordsByIds(usersCollection.Id, mapKeys(unresolved))
	if err != nil {
		return nil, err
	}

	for _, userRecord := range userRecords {
		identities[userRecord.Id] = paymentIdentity{
			Username: userRecord.GetString("username"),
			Name:     userRecord.GetString("name"),
		}
	}

	return identities, nil
}

func paymentUserIDs(paymentRecords []*pbmodels.Record) []string {
	userIDs := make([]string, 0, len(paymentRecords))
	for _, paymentRecord := range paymentRecords {
		userIDs = append(userIDs, paymentRecord.GetString("user_id"))
	}

	return userIDs
}

func mapKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	return keys
}

func uniqueNonEmptyStrings(values []string) []string {
	unique := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}

		seen[value] = struct{}{}
		unique = append(unique, value)
	}

	return unique
}
