package table

import (
	"time"
)

type ChunithmBinding struct {
	ImID     string `gorm:"column:im_id;type:varchar(30);primaryKey"`
	Platform string `gorm:"column:platform;type:varchar(20);primaryKey"`
	Server   string `gorm:"column:server;type:varchar(10);primaryKey"`
	AimeID   string `gorm:"column:aime_id;type:varchar(50);not null"`
}

func (ChunithmBinding) TableName() string {
	return "bindings"
}

type ChunithmDefaultServer struct {
	ImID     string `gorm:"column:im_id;type:varchar(30);primaryKey"`
	Platform string `gorm:"column:platform;type:varchar(20);primaryKey"`
	Server   string `gorm:"column:server;type:varchar(10);not null"`
}

func (ChunithmDefaultServer) TableName() string {
	return "defaults"
}

type ChunithmMusicAlias struct {
	ID      int64  `gorm:"column:id;primaryKey;autoIncrement"`
	MusicID int    `gorm:"column:music_id;not null"`
	Alias   string `gorm:"column:alias;type:varchar(100);not null"`
}

func (ChunithmMusicAlias) TableName() string {
	return "chunithm_aliases"
}

type ChunithmChartData struct {
	MusicID    int      `gorm:"column:music_id;primaryKey"`
	Difficulty int      `gorm:"column:difficulty;primaryKey"`
	Creator    *string  `gorm:"column:creator;type:varchar(50)"`
	BPM        *float64 `gorm:"column:bpm"`
	TapCount   *int     `gorm:"column:tap_count"`
	HoldCount  *int     `gorm:"column:hold_count"`
	SlideCount *int     `gorm:"column:slide_count"`
	AirCount   *int     `gorm:"column:air_count"`
	FlickCount *int     `gorm:"column:flick_count"`
	TotalCount *int     `gorm:"column:total_count"`
}

func (ChunithmChartData) TableName() string {
	return "chart_data"
}

type ChunithmMusic struct {
	MusicID        int        `gorm:"column:music_id;primaryKey"`
	Title          string     `gorm:"column:title;type:varchar(255);not null"`
	Artist         string     `gorm:"column:artist;type:varchar(255);not null"`
	Category       *string    `gorm:"column:category;type:varchar(50)"`
	Version        *string    `gorm:"column:version;type:varchar(10)"`
	ReleaseDate    *time.Time `gorm:"column:release_date"`
	IsDeleted      int        `gorm:"column:is_deleted;type:int;default:0;check:is_deleted IN (0,1)"`
	DeletedVersion *string    `gorm:"column:deleted_version;type:varchar(10)"`
}

func (ChunithmMusic) TableName() string {
	return "music"
}

type ChunithmMusicDifficulty struct {
	MusicID    int      `gorm:"column:music_id;primaryKey"`
	Version    string   `gorm:"column:version;type:varchar(10);primaryKey"`
	Diff0Const *float64 `gorm:"column:diff0_const;type:numeric(3,1)"`
	Diff1Const *float64 `gorm:"column:diff1_const;type:numeric(3,1)"`
	Diff2Const *float64 `gorm:"column:diff2_const;type:numeric(3,1)"`
	Diff3Const *float64 `gorm:"column:diff3_const;type:numeric(3,1)"`
	Diff4Const *float64 `gorm:"column:diff4_const;type:numeric(3,1)"`
}

func (ChunithmMusicDifficulty) TableName() string {
	return "music_difficulties"
}
