package entities

import (
	"time"

	"auth-service/src/domain/errors"
	"auth-service/src/domain/valueObjects"
)

// User represents a user domain entity
type User struct {
	id            *valueObjects.UserID
	username      *valueObjects.Username
	email         *valueObjects.Email
	password      *valueObjects.Password
	profile       *Profile
	isActive      bool
	lastLoginAt   *time.Time
	loginAttempts int
	auditInfo     *AuditInfo
}

// NewUser creates a new User entity
func NewUser(username, email, password, createdBy string) (*User, error) {
	// Create value objects
	userID, err := valueObjects.NewUserID()
	if err != nil {
		return nil, err
	}

	usernameVO, err := valueObjects.NewUsername(username)
	if err != nil {
		return nil, err
	}

	emailVO, err := valueObjects.NewEmail(email)
	if err != nil {
		return nil, err
	}

	passwordVO, err := valueObjects.NewPassword(password)
	if err != nil {
		return nil, err
	}

	auditInfo := NewAuditInfo(createdBy)

	return &User{
		id:            &userID,
		username:      &usernameVO,
		email:         &emailVO,
		password:      &passwordVO,
		isActive:      true,
		loginAttempts: 0,
		auditInfo:     &auditInfo,
	}, nil
}

// NewUserFromRepository creates a User entity from repository data
func NewUserFromRepository(
	userID valueObjects.UserID,
	username valueObjects.Username,
	email valueObjects.Email,
	password valueObjects.Password,
	isActive bool,
	lastLoginAt *time.Time,
	loginAttempts int,
	auditInfo AuditInfo,
) *User {
	return &User{
		id:            &userID,
		username:      &username,
		email:         &email,
		password:      &password,
		isActive:      isActive,
		lastLoginAt:   lastLoginAt,
		loginAttempts: loginAttempts,
		auditInfo:     &auditInfo,
	}
}

// NewUserFromCache creates a User entity from cache data
func NewUserFromCache(
	userID valueObjects.UserID,
	username valueObjects.Username,
	email valueObjects.Email,
	isActive bool,
	auditInfo AuditInfo,
) *User {
	// Create a dummy password for cache reconstruction
	// In real implementation, you might want to store a hash or omit password from cache
	dummyPassword, _ := valueObjects.NewPassword("cached_user")

	return &User{
		id:            &userID,
		username:      &username,
		email:         &email,
		password:      &dummyPassword,
		isActive:      isActive,
		lastLoginAt:   nil,
		loginAttempts: 0,
		auditInfo:     &auditInfo,
	}
}

// Business logic methods

// Activate activates the user
func (u *User) Activate() error {
	if u.isActive {
		return errors.ErrUserAlreadyActive
	}
	u.isActive = true
	u.auditInfo.Update("system")
	return nil
}

// Deactivate deactivates the user
func (u *User) Deactivate() error {
	if !u.isActive {
		return errors.ErrUserAlreadyInactive
	}
	u.isActive = false
	u.auditInfo.Update("system")
	return nil
}

// ChangePassword changes the user's password
func (u *User) ChangePassword(newPassword string) error {
	passwordVO, err := valueObjects.NewPassword(newPassword)
	if err != nil {
		return err
	}
	u.password = &passwordVO
	u.auditInfo.Update("system")
	return nil
}

// UpdateProfile updates the user's profile
func (u *User) UpdateProfile(profile *Profile) error {
	u.profile = profile
	u.auditInfo.Update("system")
	return nil
}

// RecordLogin records a successful login
func (u *User) RecordLogin() {
	now := time.Now()
	u.lastLoginAt = &now
	u.loginAttempts = 0
	u.auditInfo.Update("system")
}

// RecordFailedLogin records a failed login attempt
func (u *User) RecordFailedLogin() {
	u.loginAttempts++
	u.auditInfo.Update("system")
}

// IsLocked checks if the user account is locked due to too many failed attempts
func (u *User) IsLocked() bool {
	return u.loginAttempts >= 5
}

// Getters
func (u *User) ID() valueObjects.UserID {
	return *u.id
}

func (u *User) Username() valueObjects.Username {
	return *u.username
}

func (u *User) Email() valueObjects.Email {
	return *u.email
}

func (u *User) Password() valueObjects.Password {
	return *u.password
}

func (u *User) Profile() *Profile {
	return u.profile
}

func (u *User) IsActive() bool {
	return u.isActive
}

func (u *User) LastLoginAt() *time.Time {
	return u.lastLoginAt
}

func (u *User) LoginAttempts() int {
	return u.loginAttempts
}

func (u *User) AuditInfo() *AuditInfo {
	return u.auditInfo
}

// Profile represents user profile information
type Profile struct {
	FirstName string
	LastName  string
	Phone     string
	Address   string
	Avatar    string
	Bio       string
}

// NewProfile creates a new Profile
func NewProfile(firstName, lastName, phone, address, avatar, bio string) *Profile {
	return &Profile{
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		Address:   address,
		Avatar:    avatar,
		Bio:       bio,
	}
}

// AuditInfo represents audit information
type AuditInfo struct {
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// NewAuditInfo creates a new AuditInfo
func NewAuditInfo(createdBy string) AuditInfo {
	now := time.Now()
	return AuditInfo{
		CreatedBy: createdBy,
		CreatedAt: now,
	}
}

// Update updates the audit info (no-op since we removed ModifiedBy/ModifiedAt)
func (a *AuditInfo) Update(modifiedBy string) {
	// No-op since we removed ModifiedBy and ModifiedAt fields
}
