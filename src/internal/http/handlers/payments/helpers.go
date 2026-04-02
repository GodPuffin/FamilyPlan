package payments

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const maxPaymentNotesLength = 500

func parseForMonth(value string) string {
	if value == "" {
		return ""
	}

	forMonthDate, err := time.Parse("2006-01", value)
	if err != nil {
		return ""
	}

	return forMonthDate.Format("2006-01-02")
}

func normalizeNotes(value string) (string, error) {
	notes := strings.TrimSpace(value)
	if utf8.RuneCountInString(notes) > maxPaymentNotesLength {
		return "", fmt.Errorf("notes must be %d characters or fewer", maxPaymentNotesLength)
	}

	return notes, nil
}
