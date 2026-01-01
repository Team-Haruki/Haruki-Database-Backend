package types

import "time"

// ================= PJSK Alias Types =================

type PJSKPendingAlias struct {
	ID          int64     `json:"id"`
	AliasType   string    `json:"alias_type"`
	AliasTypeID int       `json:"alias_type_id"`
	Alias       string    `json:"alias"`
	SubmittedBy string    `json:"submitted_by"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// ================= PJSK Preference Types =================

type PJSKPreference struct {
	Option string `json:"option,omitempty"`
	Value  string `json:"value"`
}

type PJSKPreferenceResponse struct {
	Options []PJSKPreference `json:"options,omitempty"`
	Option  *PJSKPreference  `json:"option,omitempty"`
}

// ================= PJSK Binding Types =================

type PJSKBinding struct {
	ID           int    `json:"id"`
	HarukiUserID int    `json:"haruki_user_id"`
	Server       string `json:"server"`
	UserID       string `json:"user_id"`
	Visible      bool   `json:"visible"`
}

type PJSKBindingResponse struct {
	Bindings []PJSKBinding `json:"bindings,omitempty"`
	Binding  *PJSKBinding  `json:"binding,omitempty"`
}

type PJSKAddBindingResponse struct {
	BindingID int `json:"binding_id"`
}
