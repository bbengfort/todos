package todos

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// Task is the primary database structure for the todos application and represents a
// single unit of work that must be completed. Tasks are primarily described by their
// title, but can also have arbitrary text details stored alongside it. Optionally, each
// task can have a deadline, which is used for reminders and ordering. Each task is
// assigned to a user, generally the user that created the task and the task can
// optionally be assigned to a checklist. The primary modification of a task is to
// complete it (which marks it as done) or to archive it (deleting it without removal).
type Task struct {
	ID          uint       `gorm:"primary_key" json:"id,omitempty"`
	UserID      uint       `json:"-"`
	User        User       `json:"-"`
	Username    string     `gorm:"-" json:"user,omitempty"`
	Title       string     `gorm:"not null;size:255" json:"title,omitempty" binding:"required"`
	Details     string     `gorm:"not null;size:4095" json:"details,omitempty"`
	Completed   bool       `json:"completed"`
	Archived    bool       `json:"archived"`
	ChecklistID *uint      `json:"checklist,omitempty"`
	Checklist   *Checklist `json:"-"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Checklist groups related tasks so that they can be managed together. A task does not
// have to belong to a checklist, though it is recommended that all tasks are assigned
// to a list to prevent them from being stranded. Checklists are owned by individual
// users, which manage their lists. Similar to tasks, checklists are described by a
// title and optional details. However, checklists can only be "completed" if all of its
// tasks are either completed or archived, and this is not directly stored in the
// database, but is rather computed on demand. Checklists can also have a deadline,
// which is used for reminders and checklist ordering.
type Checklist struct {
	ID        uint       `gorm:"primary_key" json:"id,omitempty"`
	UserID    uint       `json:"-"`
	User      User       `json:"-"`
	Username  string     `gorm:"-" json:"user,omitempty"`
	Title     string     `gorm:"not null;size:255" json:"title,omitempty"`
	Details   string     `gorm:"not null;size:4095" json:"details,omitempty"`
	Completed uint       `gorm:"-" json:"completed,omitempty"`
	Archived  uint       `gorm:"-" json:"archived,omitempty"`
	Size      uint       `gorm:"-" json:"size"`
	Deadline  *time.Time `json:"deadline,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Tasks     []Task     `json:"tasks,omitempty"`
}

// User is primarily used for authentication and storing json web tokens. Each user in
// the system manages their own tasks and checklists through the API. This is the
// primary partitioning mechanism between tasks.
type User struct {
	ID            uint        `gorm:"primary_key" json:"id"`
	Username      string      `gorm:"unique;not null;size:255" json:"username"`
	Email         string      `gorm:"unique;not null;size:255" json:"email"`
	Password      string      `gorm:"not null;size:255" json:"-"`
	IsAdmin       bool        `json:"is_admin"`
	DefaultListID *uint       `json:"default_checklist,omitempty"`
	DefaultList   *Checklist  `json:"-"`
	LastSeen      *time.Time  `json:"last_seen"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	Tasks         []Task      `json:"-"`
	Lists         []Checklist `json:"-"`
	Tokens        []Token     `json:"-"`
}

// Token holds an access and refresh tokens, which are granted after authentication and
// used to authorize further requests using a Bearer header. The refresh token is used
// to update authentication without having to submit a login and password again.
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

	// Migrate todos models
	db.AutoMigrate(&Task{}, &Checklist{})
	db.Model(&Task{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	db.Model(&Task{}).AddForeignKey("checklist_id", "checklists(id)", "CASCADE", "RESTRICT")
	db.Model(&Checklist{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	db.Model(&User{}).AddForeignKey("default_list_id", "checklists(id)", "CASCADE", "RESTRICT")

	errors := db.GetErrors()
	if len(errors) > 1 {
		return fmt.Errorf("%d errors occurred during migration", len(errors))
	}

	if len(errors) == 1 {
		return errors[0]
	}
	return nil
}
