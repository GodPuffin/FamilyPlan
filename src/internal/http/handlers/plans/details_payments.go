package plans

import (
	"fmt"

	"familyplan/src/internal/domain"

	"github.com/pocketbase/pocketbase"
)

func loadPendingPayments(app *pocketbase.PocketBase, planID string) ([]domain.Payment, error) {
	return loadPaymentsByFilter(app, planID, fmt.Sprintf("plan_id = '%s' && status = 'pending'", planID))
}

func loadAllPayments(app *pocketbase.PocketBase, planID string) ([]domain.Payment, error) {
	return loadPaymentsByFilter(app, planID, fmt.Sprintf("plan_id = '%s'", planID))
}

func loadUserPayments(app *pocketbase.PocketBase, planID, userID string) ([]domain.Payment, error) {
	paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
	if err != nil {
		return nil, err
	}

	paymentRecords, err := app.Dao().FindRecordsByFilter(
		paymentsCollection.Id,
		fmt.Sprintf("plan_id = '%s' && user_id = '%s'", planID, userID),
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

func loadPaymentsByFilter(app *pocketbase.PocketBase, planID, filter string) ([]domain.Payment, error) {
	paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
	if err != nil {
		return nil, err
	}

	paymentRecords, err := app.Dao().FindRecordsByFilter(
		paymentsCollection.Id,
		filter,
		"-created",
		100,
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

	artificialMembership, _ := app.Dao().FindFirstRecordByFilter(
		membershipsCollection.Id,
		fmt.Sprintf("plan_id = '%s' && user_id = '%s' && is_artificial = true", planID, userID),
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
