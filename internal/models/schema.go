package schema

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	Email     string    `json:"email" gorm:"size:50;uniqueIndex;not null"`
	Password  string    `json:"password" gorm:"size:20;not null"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	Salt      string    `json:"salt" gorm:"size:255;not null"`
	CreatedAt time.Time
}

type CustomDomain struct {
	Id          int64  `json:"id" gorm:"primaryKey"`
	OwnerId     string `json:"owner_id" gorm:"type:uuid;not null"`
	Owner       User   `gorm:"constraint:OnDelete:CASCADE"`
	Domain      string `gorm:"type:citext;not null;uniqueIndex"`
	IsVerified  bool   `gorm:"not null;default:false"`
	VerifyToken string
	CreatedAt   time.Time
}

type ShortLink struct {
	Id        uint64       `gorm:"primaryKey;autoIncrement"`
	OwnerId   uuid.UUID    `gorm:"type:uuid;not null;index"`
	Owner     User         `gorm:"constraint:OnDelete:CASCADE"`
	DomainId  *uint64      `gorm:"index"`
	Domain    CustomDomain `gorm:"constraint:OnDelete:SET NULL"`
	Code      string       `gorm:"type:text;not null;index:idx_lookup,priority:1"`
	TargetURL string       `gorm:"type:text;not null"`
	IsActive  bool         `gorm:"not null;default:true;index"`
	CreatedAt time.Time    `grom:"not null;default:now()"`
	ExpiredAt *time.Time   `gorm:"index"`
	MaxClick  *int
	Notes     *string
}

// GORM v2 way to define composite unique indexes via tags:
type _ShortLinkIndexes struct {
	DomainID *uint64 `gorm:"uniqueIndex:ux_domain_code"`
	Code     string  `gorm:"uniqueIndex:ux_domain_code"`
}

// We embed a phantom struct to apply tags; GORM reads tags from all fields.
// (Alternative: use db.AutoMigrate then db.Exec for indexes.)
type ShortLinkWithIndexes struct {
	ShortLink
	_ShortLinkIndexes _ShortLinkIndexes `gorm:"-:all"` // ignored at runtime
}
