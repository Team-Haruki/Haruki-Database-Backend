// Package types provides Chunithm-specific types and response structures.
package types

import "time"

// ================= Chunithm Music Types =================

// ChunithmMusicInfo represents basic music information
type ChunithmMusicInfo struct {
	MusicID        int        `json:"music_id"`
	Title          string     `json:"title"`
	Artist         string     `json:"artist"`
	Category       *string    `json:"category,omitempty"`
	Version        *string    `json:"version,omitempty"`
	ReleaseDate    *time.Time `json:"release_date,omitempty"`
	IsDeleted      *bool      `json:"is_deleted,omitempty"`
	DeletedVersion *string    `json:"deleted_version,omitempty"`
}

// ChunithmMusicDifficulty represents music difficulty constants
type ChunithmMusicDifficulty struct {
	MusicID int      `json:"music_id"`
	Version string   `json:"version"`
	Diff0   *float64 `json:"diff0_const,omitempty"`
	Diff1   *float64 `json:"diff1_const,omitempty"`
	Diff2   *float64 `json:"diff2_const,omitempty"`
	Diff3   *float64 `json:"diff3_const,omitempty"`
	Diff4   *float64 `json:"diff4_const,omitempty"`
}

// ChunithmChartData represents chart data for a specific difficulty
type ChunithmChartData struct {
	Difficulty int      `json:"difficulty"`
	Creator    *string  `json:"creator,omitempty"`
	BPM        *float64 `json:"bpm,omitempty"`
	TapCount   *int     `json:"tap_count,omitempty"`
	HoldCount  *int     `json:"hold_count,omitempty"`
	SlideCount *int     `json:"slide_count,omitempty"`
	AirCount   *int     `json:"air_count,omitempty"`
	FlickCount *int     `json:"flick_count,omitempty"`
	TotalCount *int     `json:"total_count,omitempty"`
}

// ChunithmMusicBatchItem represents a batch item for music operations
type ChunithmMusicBatchItem struct {
	Version    *string           `json:"version,omitempty"`
	Difficulty []*float64        `json:"difficulty"`
	Info       ChunithmMusicInfo `json:"info"`
}

// ================= Chunithm Binding Types =================

// ChunithmDefaultServer represents a user's default server setting
type ChunithmDefaultServer struct {
	UserID int    `json:"user_id"`
	Server string `json:"server"`
}

// ChunithmBinding represents a user's binding
type ChunithmBinding struct {
	UserID int     `json:"user_id"`
	Server *string `json:"server,omitempty"`
	AimeID *string `json:"aime_id,omitempty"`
}

// ================= Chunithm Alias Types =================

// ChunithmMusicAlias represents a music alias
type ChunithmMusicAlias struct {
	ID    int64  `json:"id,omitempty"`
	Alias string `json:"alias"`
}
