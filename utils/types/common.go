// Package types provides common types and response structures.
package types

import "time"

// ================= Common Response Types =================

// AliasToIDResponse is the response for alias to ID lookup
type AliasToIDResponse struct {
	MatchIDs []int `json:"match_ids"`
}

// AliasListResponse is the response for listing aliases
type AliasListResponse struct {
	Aliases []string `json:"aliases"`
}

// ================= Common Request Types =================

// AliasRequest is the request for alias operations
type AliasRequest struct {
	Alias string `json:"alias"`
}

// RejectRequest is the request for rejecting an alias
type RejectRequest struct {
	Reason string `json:"reason"`
}

// ================= Common Schema Types =================

// PendingAlias represents a pending alias submission
type PendingAlias struct {
	ID          int64     `json:"id"`
	AliasType   string    `json:"alias_type"`
	AliasTypeID int       `json:"alias_type_id"`
	Alias       string    `json:"alias"`
	SubmittedBy string    `json:"submitted_by"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// RejectedAlias represents a rejected alias
type RejectedAlias struct {
	ID          int64  `json:"id"`
	AliasType   string `json:"alias_type"`
	AliasTypeID int    `json:"alias_type_id"`
	Alias       string `json:"alias"`
	Submitter   string `json:"submitter"`
	Reason      string `json:"reason"`
}
