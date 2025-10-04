package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// User represents a user in the system.
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"not null" json:"-"` // "-" excludes from JSON
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Location  string    `json:"location"` // City, State for now
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	ShowAttendances []ShowAttendance `gorm:"foreignKey:UserID" json:"show_attendances,omitempty"`
}

// Band represents a musical band/artist.
type Band struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Name        string    `gorm:"not null;index" json:"name"`
	Genre       string    `json:"genre"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Shows []Show `gorm:"foreignKey:BandID" json:"shows,omitempty"`
}

// Venue represents a concert venue
type Venue struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Name      string    `gorm:"not null;index" json:"name"`
	City      string    `gorm:"not null" json:"city"`
	State     string    `gorm:"not null" json:"state"`
	Country   string    `gorm:"default:'USA'" json:"country"`
	Address   string    `json:"address"`
	Capacity  int       `json:"capacity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Shows []Show `gorm:"foreignKey:VenueID" json:"shows,omitempty"`
}

// Show represents a concert/performance
type Show struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	BandID    uuid.UUID `gorm:"type:uuid;not null;index" json:"band_id"`
	VenueID   uuid.UUID `gorm:"type:uuid;not null;index" json:"venue_id"`
	Date      time.Time `gorm:"not null;index" json:"date"`
	SetlistID string    `json:"setlist_id,omitempty"` // For setlist.fm integration later
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Band            Band             `gorm:"foreignKey:BandID" json:"band"`
	Venue           Venue            `gorm:"foreignKey:VenueID" json:"venue"`
	ShowAttendances []ShowAttendance `gorm:"foreignKey:ShowID" json:"attendances,omitempty"`
}

// ShowAttendance represents a user's attendance at a show
type ShowAttendance struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	ShowID       uuid.UUID `gorm:"type:uuid;not null;index" json:"show_id"`
	Rating       *float64  `gorm:"check:rating >= 1 AND rating <= 5" json:"rating"` // 1-5 scale, nullable
	FavoriteSong string    `json:"favorite_song"`
	Notes        string    `json:"notes"`
	Attended     bool      `gorm:"default:true" json:"attended"` // For future "want to attend" feature
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user"`
	Show Show `gorm:"foreignKey:ShowID" json:"show"`
}

// BeforeCreate hooks for generating UUIDs
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}

func (b *Band) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return
}

func (v *Venue) BeforeCreate(tx *gorm.DB) (err error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return
}

func (s *Show) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

func (sa *ShowAttendance) BeforeCreate(tx *gorm.DB) (err error) {
	if sa.ID == uuid.Nil {
		sa.ID = uuid.New()
	}
	return
}
