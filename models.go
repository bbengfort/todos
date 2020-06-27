package todos

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// Todo is the primary data structure for interacting with the application. Note that
// while a Todo must be assigned to a User, the data must be manually updated with the
// UserEmail for serialization (rather than nesting the User object).
type Todo struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	UserID    uint       `json:"-"`
	User      User       `json:"-"`
	UserEmail string     `gorm:"-" json:"user"`
	Title     string     `gorm:"not null;size:255" json:"title" binding:"required"`
	Details   string     `gorm:"not null;size:4095" json:"details"`
	Completed bool       `json:"completed"`
	Archived  bool       `json:"archived"`
	ListID    *uint      `json:"list,omitempty"`
	List      *List      `json:"-"`
	Deadline  *time.Time `json:"deadline,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// List groups related Todos so that they can be managed together. A Todo does not have
// to belong to a list, and there is no "default" list except for any Todo that is not
// assigned to a list, e.g. the field is nullable. Lists are owned by users, but they
// are described by the UserEmail which must be manually updated for serialization.
// Completed, Archived, and Size are used to compute the percentage complete/archived
// based on the Size (number of items) of the list. These are not stored on the database
// but are computed from the associated Todos and serialized in response.
type List struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	UserID    uint       `json:"-"`
	User      User       `json:"-"`
	UserEmail string     `gorm:"-" json:"user"`
	Title     string     `gorm:"not null;size:255" json:"title"`
	Details   string     `gorm:"not null;size:4095" json:"details"`
	Completed uint       `gorm:"-" json:"completed,omitempty"`
	Archived  uint       `gorm:"-" json:"archived,omitempty"`
	Size      uint       `gorm:"-" json:"size"`
	Deadline  *time.Time `json:"deadline,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Todos     []Todo     `json:"todos,omitempty"`
}

// User is primarily used for authentication and storing json web tokens.
type User struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	Username  string     `gorm:"unique;not null;size:255" json:"username"`
	Email     string     `gorm:"unique;not null;size:255" json:"email"`
	Password  string     `gorm:"not null;size:255" json:"-"`
	IsAdmin   bool       `json:"is_admin"`
	LastSeen  *time.Time `json:"last_seen"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Todos     []Todo     `json:"-"`
	Lists     []List     `json:"-"`
	Tokens    []Token    `json:"-"`
}

// Token holds an access and refresh tokens.
type Token struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	User         User      `json:"-"`
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshBy    time.Time `json:"refresh_by"`
	accessToken  string
	refreshToken string
}

// Migrate the schema based on the models defined below.
func Migrate(db *gorm.DB) (err error) {
	// Migrate auth models
	db.AutoMigrate(&User{}, &Token{})
	db.Model(&Token{}).AddForeignKey("user_id", "users(id)", "CASCADE", "RESTRICT")

	// Migrate todo models
	db.AutoMigrate(&Todo{}, &List{})
	db.Model(&Todo{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	db.Model(&Todo{}).AddForeignKey("list_id", "lists(id)", "CASCADE", "RESTRICT")
	db.Model(&List{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")

	errors := db.GetErrors()
	if len(errors) > 1 {
		return fmt.Errorf("%d errors occurred during migration", len(errors))
	}

	if len(errors) == 1 {
		return errors[0]
	}
	return nil
}
