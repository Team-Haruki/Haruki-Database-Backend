// Package types provides PJSK-specific types and response structures.
package types

import "time"

// ================= PJSK Alias Types =================

// PJSKPendingAlias represents a pending PJSK alias submission
type PJSKPendingAlias struct {
	ID          int64     `json:"id"`
	AliasType   string    `json:"alias_type"`
	AliasTypeID int       `json:"alias_type_id"`
	Alias       string    `json:"alias"`
	SubmittedBy string    `json:"submitted_by"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// ================= PJSK Preference Types =================

// PJSKPreference represents a user preference for PJSK
type PJSKPreference struct {
	Option string `json:"option"`
	Value  string `json:"value"`
}

// PJSKPreferenceResponse is the response for preference queries
type PJSKPreferenceResponse struct {
	Options []PJSKPreference `json:"options,omitempty"`
	Option  *PJSKPreference  `json:"option,omitempty"`
}

// ================= PJSK Binding Types =================

// PJSKBinding represents a PJSK binding
type PJSKBinding struct {
	ID           int    `json:"id"`
	HarukiUserID int    `json:"haruki_user_id"`
	Server       string `json:"server"`
	UserID       string `json:"user_id"`
	Visible      bool   `json:"visible"`
}

// PJSKBindingResponse is the response for binding queries
type PJSKBindingResponse struct {
	Bindings []PJSKBinding `json:"bindings,omitempty"`
	Binding  *PJSKBinding  `json:"binding,omitempty"`
}

// PJSKAddBindingResponse is the response for adding a binding
type PJSKAddBindingResponse struct {
	BindingID int `json:"binding_id"`
}
