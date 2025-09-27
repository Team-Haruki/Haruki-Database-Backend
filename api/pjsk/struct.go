package pjsk

import "time"

type AliasToObjectIdResponse struct {
	MatchIDs []int `json:"match_ids"`
}

type AllAliasesResponse struct {
	Aliases []string `json:"aliases"`
}

type PendingAlias struct {
	ID          int64     `json:"id"`
	AliasType   string    `json:"alias_type"`
	AliasTypeID int       `json:"alias_type_id"`
	Alias       string    `json:"alias"`
	SubmittedBy string    `json:"submitted_by"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type PendingAliasListResponse struct {
	Rows    int            `json:"rows"`
	Results []PendingAlias `json:"results"`
}

type RejectedAliasResponse struct {
	ID          int    `json:"id"`
	AliasType   string `json:"alias_type"`
	AliasTypeID int    `json:"alias_type_id"`
	Alias       string `json:"alias"`
	Submitter   string `json:"submitter"`
	Reason      string `json:"reason"`
}

type AliasRequest struct {
	Alias string `json:"alias"`
}

type RejectRequest struct {
	Reason string `json:"reason"`
}

type UserPreferenceSchema struct {
	Option string `json:"option"`
	Value  string `json:"value"`
}

type UserPreferenceResponse struct {
	Options []UserPreferenceSchema `json:"options,omitempty"`
	Option  *UserPreferenceSchema  `json:"option,omitempty"`
}

type BindingSchema struct {
	ID       int    `json:"id"`
	Platform string `json:"platform"`
	ImID     string `json:"im_id"`
	Server   string `json:"server"`
	UserID   string `json:"user_id"`
	Visible  bool   `json:"visible"`
}

type BindingResponse struct {
	Bindings []BindingSchema `json:"bindings,omitempty"`
	Binding  *BindingSchema  `json:"binding,omitempty"`
}

type AddBindingSuccessResponse struct {
	BindingID int `json:"binding_id"`
}
