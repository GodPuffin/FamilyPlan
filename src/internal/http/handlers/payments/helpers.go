package payments

import "time"

func parseForMonth(value string) string {
	if value == "" {
		return ""
	}

	forMonthDate, err := time.Parse("2006-01", value)
	if err != nil {
		return ""
	}

	return forMonthDate.Format(time.RFC3339)
}
