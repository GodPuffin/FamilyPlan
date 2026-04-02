package plans

import (
	"familyplan/src/internal/domain"
	"familyplan/src/internal/planutil"

	"github.com/pocketbase/pocketbase"
)

func loadPendingPayments(app *pocketbase.PocketBase, planID string) ([]domain.Payment, error) {
	return loadPaymentsByTerms(app, planID,
		planutil.FilterTerm{Field: "plan_id", Value: planID},
		planutil.FilterTerm{Field: "status", Value: "pending"},
	)
}

func loadAllPayments(app *pocketbase.PocketBase, planID string) ([]domain.Payment, error) {
	return loadPaymentsByTerms(app, planID,
		planutil.FilterTerm{Field: "plan_id", Value: planID},
	)
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
		filter,
		"-created",
		20,
		0,
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

func loadPaymentsByTerms(app *pocketbase.PocketBase, planID string, terms ...planutil.FilterTerm) ([]domain.Payment, error) {
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
		filter,
		"-created",
		-1,
		0,
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
	)
	if err != nil {
		return "", "", err
	}

	artificialMembership, _ := app.Dao().FindFirstRecordByFilter(
		membershipsCollection.Id,
		filter+" && is_artificial = true",
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
