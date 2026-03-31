package model

type UserRole string

const (
	RoleUser      UserRole = "user"
	RoleModerator UserRole = "moderator"
)

type User struct {
	ID       uint     `gorm:"primaryKey"`
	Username string   `gorm:"uniqueIndex;not null"`
	Password string   `gorm:"not null"`
	Role     UserRole `gorm:"not null;default:'user'"`
}
