package main

// SessionData holds user session information
type SessionData struct {
	IsAuthenticated bool
	UserId          string
	Username        string
	Name            string
}

// FamilyPlan represents a subscription plan that can be shared among family/friends
type FamilyPlan struct {
	Id             string  `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Cost           float64 `json:"cost"`
	IndividualCost float64 `json:"individual_cost"`
	Owner          string  `json:"owner"`
	JoinCode       string  `json:"join_code"`
	CreatedAt      string  `json:"created_at"`
	MembersCount   int     `json:"members_count"`
	Balance        float64 `json:"balance"`
}

// Member represents a user who is part of a family plan
type Member struct {
	Id             string  `json:"id"`
	Username       string  `json:"username"`
	Name           string  `json:"name"`
	Balance        float64 `json:"balance"`
	LeaveRequested bool    `json:"leave_requested"`
	DateEnded      string  `json:"date_ended"`
	IsArtificial   bool    `json:"is_artificial"`
}

// JoinRequest represents a user's request to join a family plan
type JoinRequest struct {
	UserId      string `json:"user_id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	RequestedAt string `json:"requested_at"`
}

// Payment represents a payment made by a member for a family plan
type Payment struct {
	Id       string  `json:"id"`
	PlanId   string  `json:"plan_id"`
	UserId   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Date     string  `json:"date"`
	Status   string  `json:"status"`
	Notes    string  `json:"notes"`
	ForMonth string  `json:"for_month"`
	Username string  `json:"username"`
	Name     string  `json:"name"`
}
