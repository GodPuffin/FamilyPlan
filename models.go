package main

// SessionData holds user session information
type SessionData struct {
	IsAuthenticated bool
	UserId          string
	Username        string
}

// FamilyPlan represents a subscription plan that can be shared among family/friends
type FamilyPlan struct {
	Id           string  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Cost         float64 `json:"cost"`
	Owner        string  `json:"owner"`
	JoinCode     string  `json:"join_code"`
	CreatedAt    string  `json:"created_at"`
	MembersCount int     `json:"members_count"`
}

// Member represents a user who is part of a family plan
type Member struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

// JoinRequest represents a user's request to join a family plan
type JoinRequest struct {
	UserId      string `json:"user_id"`
	Username    string `json:"username"`
	RequestedAt string `json:"requested_at"`
}
