package plans

import (
	"familyplan/src/internal/domain"
	"familyplan/src/internal/planutil"

	"github.com/pocketbase/pocketbase"
)

func loadPendingPayments(app *pocketbase.PocketBase, planID string) ([]domain.Payment, error) {
	return loadPaymentsByTerms(app, planID, -1, 0,
		planutil.FilterTerm{Field: "plan_id", Value: planID},
		planutil.FilterTerm{Field: "status", Value: "pending"},
	)
}

func loadAllPaymentsPage(app *pocketbase.PocketBase, planID string, page, pageSize int) ([]domain.Payment, map[string]interface{}, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = memberPaymentsPageSize
	}

	payments, err := loadPaymentsByTerms(app, planID, pageSize+1, (page-1)*pageSize,
		planutil.FilterTerm{Field: "plan_id", Value: planID},
	)
	if err != nil {
		return nil, nil, err
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

	payments := make([]domain.Payment, 0, len(paymentRecords))
	for _, paymentRecord := range paymentRecords {
		username, name, err := paymentIdentity(app, planID, paymentRecord.GetString("user_id"))
		if err != nil {
			continue
		}

		payments = append(payments, buildPayment(paymentRecord, username, name))
	}

	return payments, nil
}

func paymentIdentity(app *pocketbase.PocketBase, planID, userID string) (string, string, error) {
	membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
	if err != nil {
		return "", "", err
	}

	filter, err := planutil.BuildEqualsFilter(
		planutil.FilterTerm{Field: "plan_id", Value: planID},
		planutil.FilterTerm{Field: "user_id", Value: userID},
		planutil.FilterTerm{Field: "is_artificial", Value: true},
	)
	if err != nil {
		return "", "", err
	}

	artificialMembership, _ := app.Dao().FindFirstRecordByFilter(
		membershipsCollection.Id,
		filter.Expression,
		filter.Params,
	)
	if artificialMembership != nil {
		return "", artificialMembership.GetString("name"), nil
	}

	usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		return "", "", err
	}

	userRecord, err := app.Dao().FindRecordById(usersCollection.Id, userID)
	if err != nil || userRecord == nil {
		return "", "", err
	}

	return userRecord.GetString("username"), userRecord.GetString("name"), nil
}
