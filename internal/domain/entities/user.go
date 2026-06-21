package entities

import (
	"errors"
	"time"

	"github.com/keepguard/bff-core/internal/domain/valueobjects"
)

// User representa um usuário no domínio
type User struct {
	id        string
	codeUser  string
	name      string
	email     valueobjects.Email
	phone     valueobjects.Phone
	companyID string
	userType  valueobjects.UserType
	status    UserStatus
	createdAt time.Time
	updatedAt time.Time
}

// UserStatus representa o status do usuário
type UserStatus string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
	UserStatusPending  UserStatus = "PENDING"
	UserStatusBlocked  UserStatus = "BLOCKED"
)

// NewUser cria um novo usuário
func NewUser(name string, email valueobjects.Email, phone valueobjects.Phone, companyID string, userType valueobjects.UserType) (*User, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	if companyID == "" {
		return nil, errors.New("companyID cannot be empty")
	}

	now := time.Now()
	return &User{
		id:        generateID(),
		codeUser:  generateCodeUser(),
		name:      name,
		email:     email,
		phone:     phone,
		companyID: companyID,
		userType:  userType,
		status:    UserStatusActive,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Getters
func (u *User) ID() string {
	return u.id
}

func (u *User) CodeUser() string {
	return u.codeUser
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Email() valueobjects.Email {
	return u.email
}

func (u *User) Phone() valueobjects.Phone {
	return u.phone
}

func (u *User) CompanyID() string {
	return u.companyID
}

func (u *User) UserType() valueobjects.UserType {
	return u.userType
}

func (u *User) Status() UserStatus {
	return u.status
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// Business Methods
func (u *User) Activate() {
	u.status = UserStatusActive
	u.updatedAt = time.Now()
}

func (u *User) Deactivate() {
	u.status = UserStatusInactive
	u.updatedAt = time.Now()
}

func (u *User) Block() {
	u.status = UserStatusBlocked
	u.updatedAt = time.Now()
}

func (u *User) SetPending() {
	u.status = UserStatusPending
	u.updatedAt = time.Now()
}

func (u *User) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	u.name = name
	u.updatedAt = time.Now()
	return nil
}

func (u *User) UpdateEmail(email valueobjects.Email) error {
	u.email = email
	u.updatedAt = time.Now()
	return nil
}

func (u *User) UpdatePhone(phone valueobjects.Phone) error {
	u.phone = phone
	u.updatedAt = time.Now()
	return nil
}

func (u *User) UpdateUserType(userType valueobjects.UserType) error {
	u.userType = userType
	u.updatedAt = time.Now()
	return nil
}

func (u *User) IsActive() bool {
	return u.status == UserStatusActive
}

func (u *User) IsBlocked() bool {
	return u.status == UserStatusBlocked
}

func (u *User) IsPending() bool {
	return u.status == UserStatusPending
}

func (u *User) CanBeActivated() bool {
	return u.status == UserStatusInactive || u.status == UserStatusPending
}

func (u *User) CanBeBlocked() bool {
	return u.status == UserStatusActive
}

// generateID gera um ID único para o usuário
func generateID() string {
	// Implementação simplificada - em produção usar UUID
	return "usr_" + time.Now().Format("20060102150405") + "_" + randomString(6)
}

// generateCodeUser gera um código único para o usuário
func generateCodeUser() string {
	// Implementação simplificada - em produção usar algoritmo mais robusto
	return "USR" + time.Now().Format("20060102150405") + randomString(4)
}
