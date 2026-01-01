package types

import "time"

// ================= Common Response Types =================

type AliasToIDResponse struct {
	MatchIDs []int `json:"match_ids"`
}

type AliasListResponse struct {
	Aliases []string `json:"aliases"`
}

// ================= Common Request Types =================

type AliasRequest struct {
	Alias string `json:"alias"`
}

type RejectRequest struct {
	Reason string `json:"reason"`
}

// ================= Common Schema Types =================

type PendingAlias struct {
	ID          int64     `json:"id"`
	AliasType   string    `json:"alias_type"`
	AliasTypeID int       `json:"alias_type_id"`
	Alias       string    `json:"alias"`
	SubmittedBy string    `json:"submitted_by"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type RejectedAlias struct {
	ID          int64  `json:"id"`
	AliasType   string `json:"alias_type"`
	AliasTypeID int    `json:"alias_type_id"`
	Alias       string `json:"alias"`
	Submitter   string `json:"submitter"`
	Reason      string `json:"reason"`
}
