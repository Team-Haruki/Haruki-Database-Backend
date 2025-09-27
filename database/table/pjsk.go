package table

import "time"

type UserBinding struct {
	ID       int    `gorm:"column:id;primaryKey;autoIncrement"`
	Platform string `gorm:"column:platform;type:varchar(20);not null;uniqueIndex:uq_user_binding"`
	ImID     string `gorm:"column:im_id;type:varchar(30);not null;index;uniqueIndex:uq_user_binding"`
	UserID   string `gorm:"column:user_id;type:varchar(30);not null;uniqueIndex:uq_user_binding"`
	Server   string `gorm:"column:server;type:varchar(2);not null;uniqueIndex:uq_user_binding"`
	Visible  bool   `gorm:"column:visible;default:true"`

	DefaultRefs []UserDefaultBinding `gorm:"foreignKey:BindingID;constraint:OnDelete:CASCADE"`
}

func (UserBinding) TableName() string {
	return "user_bindings"
}

type UserDefaultBinding struct {
	ID        int    `gorm:"column:id;primaryKey;autoIncrement"`
	ImID      string `gorm:"column:im_id;type:varchar(30);not null;uniqueIndex:uq_user_default_binding"`
	Platform  string `gorm:"column:platform;type:varchar(20);not null;uniqueIndex:uq_user_default_binding"`
	Server    string `gorm:"column:server;type:varchar(7);not null;uniqueIndex:uq_user_default_binding"`
	BindingID int    `gorm:"column:binding_id;not null;index"`

	Binding UserBinding `gorm:"foreignKey:BindingID;references:ID;constraint:OnDelete:CASCADE"`
}

func (UserDefaultBinding) TableName() string {
	return "user_default_bindings"
}

type UserPreference struct {
	ImID     string `gorm:"column:im_id;type:varchar(30);primaryKey"`
	Platform string `gorm:"column:platform;type:varchar(20);primaryKey"`
	Option   string `gorm:"column:option;type:varchar(50);primaryKey"`
	Value    string `gorm:"column:value;type:varchar(50);not null"`
}

func (UserPreference) TableName() string {
	return "user_preferences"
}

type Alias struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement"`
	AliasType   string `gorm:"column:alias_type;type:varchar(20);not null"`
	AliasTypeID int    `gorm:"column:alias_type_id;not null"`
	Alias       string `gorm:"column:alias;type:varchar(100);not null"`
}

func (Alias) TableName() string {
	return "aliases"
}

type PendingAlias struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	AliasType   string    `gorm:"column:alias_type;type:varchar(20);not null"`
	AliasTypeID int       `gorm:"column:alias_type_id;not null"`
	Alias       string    `gorm:"column:alias;type:varchar(100);not null"`
	SubmittedBy string    `gorm:"column:submitted_by;type:varchar(100);not null"`
	SubmittedAt time.Time `gorm:"column:submitted_at;not null"`
}

func (PendingAlias) TableName() string {
	return "pending_aliases"
}

type RejectedAlias struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	AliasType   string    `gorm:"column:alias_type;type:varchar(20);not null"`
	AliasTypeID int       `gorm:"column:alias_type_id;not null"`
	Alias       string    `gorm:"column:alias;type:varchar(100);not null"`
	ReviewedBy  string    `gorm:"column:reviewed_by;type:varchar(100);not null"`
	Reason      string    `gorm:"column:reason;type:varchar(255);not null"`
	ReviewedAt  time.Time `gorm:"column:reviewed_at;not null"`
}

func (RejectedAlias) TableName() string {
	return "rejected_aliases"
}

type GroupAlias struct {
	GroupID     string `gorm:"column:group_id;type:varchar(50);primaryKey"`
	AliasType   string `gorm:"column:alias_type;type:varchar(20);primaryKey"`
	AliasTypeID int    `gorm:"column:alias_type_id;primaryKey"`
	Alias       string `gorm:"column:alias;type:varchar(100);primaryKey"`
}

func (GroupAlias) TableName() string {
	return "group_aliases"
}

type AliasAdmin struct {
	ImID string `gorm:"column:im_id;type:varchar(100);primaryKey"`
	Name string `gorm:"column:name;type:varchar(100);not null"`
}

func (AliasAdmin) TableName() string {
	return "alias_admins"
}
