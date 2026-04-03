package plans

import "strconv"

const (
	memberPaymentsPageParam = "member_payments_page"
	memberPaymentsPageSize  = 10
)

func memberPaymentsPage(raw string) int {
	page, err := strconv.Atoi(raw)
	if err != nil || page < 1 {
		return 1
	}

	return page
}

func buildMemberPaymentsPagination(page int, hasNext bool) map[string]interface{} {
	if page < 1 {
		page = 1
	}

	prevPage := 1
	if page > 1 {
		prevPage = page - 1
	}

	return map[string]interface{}{
		"CurrentPage": page,
		"HasPrev":     page > 1,
		"PrevPage":    prevPage,
		"HasNext":     hasNext,
		"NextPage":    page + 1,
	}
}
