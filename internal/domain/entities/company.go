package entities

import (
	"errors"
	"time"

	"github.com/keepguard/bff-core/internal/domain/valueobjects"
)

// Company representa uma empresa no domínio
type Company struct {
	id           string
	tenantId string
	name         string
	legalName    string
	cnpj         valueobjects.CNPJ
	status       CompanyStatus
	createdAt    time.Time
	updatedAt    time.Time
}

// CompanyStatus representa o status da empresa
type CompanyStatus string

const (
	CompanyStatusActive   CompanyStatus = "ACTIVE"
	CompanyStatusInactive CompanyStatus = "INACTIVE"
	CompanyStatusPending  CompanyStatus = "PENDING"
	CompanyStatusBlocked  CompanyStatus = "BLOCKED"
)

// NewCompany cria uma nova empresa
func NewCompany(tenantId, name, legalName string, cnpj valueobjects.CNPJ) (*Company, error) {
	if tenantId == "" {
		return nil, errors.New("tenantId cannot be empty")
	}

	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	if legalName == "" {
		return nil, errors.New("legalName cannot be empty")
	}

	now := time.Now()
	return &Company{
		id:           generateCompanyID(),
		tenantId: tenantId,
		name:         name,
		legalName:    legalName,
		cnpj:         cnpj,
		status:       CompanyStatusActive,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// Getters
func (c *Company) ID() string {
	return c.id
}

func (c *Company) TenantId() string {
	return c.tenantId
}

func (c *Company) Name() string {
	return c.name
}

func (c *Company) LegalName() string {
	return c.legalName
}

func (c *Company) CNPJ() valueobjects.CNPJ {
	return c.cnpj
}

func (c *Company) Status() CompanyStatus {
	return c.status
}

func (c *Company) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Company) UpdatedAt() time.Time {
	return c.updatedAt
}

// Business Methods
func (c *Company) Activate() {
	c.status = CompanyStatusActive
	c.updatedAt = time.Now()
}

func (c *Company) Deactivate() {
	c.status = CompanyStatusInactive
	c.updatedAt = time.Now()
}

func (c *Company) Block() {
	c.status = CompanyStatusBlocked
	c.updatedAt = time.Now()
}

func (c *Company) SetPending() {
	c.status = CompanyStatusPending
	c.updatedAt = time.Now()
}

func (c *Company) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	c.name = name
	c.updatedAt = time.Now()
	return nil
}

func (c *Company) UpdateLegalName(legalName string) error {
	if legalName == "" {
		return errors.New("legalName cannot be empty")
	}
	c.legalName = legalName
	c.updatedAt = time.Now()
	return nil
}

func (c *Company) UpdateCNPJ(cnpj valueobjects.CNPJ) error {
	c.cnpj = cnpj
	c.updatedAt = time.Now()
	return nil
}

func (c *Company) IsActive() bool {
	return c.status == CompanyStatusActive
}

func (c *Company) IsBlocked() bool {
	return c.status == CompanyStatusBlocked
}

func (c *Company) IsPending() bool {
	return c.status == CompanyStatusPending
}

func (c *Company) CanBeActivated() bool {
	return c.status == CompanyStatusInactive || c.status == CompanyStatusPending
}

func (c *Company) CanBeBlocked() bool {
	return c.status == CompanyStatusActive
}

func (c *Company) CanCreateUsers() bool {
	return c.status == CompanyStatusActive
}

// generateCompanyID gera um ID único para a empresa
func generateCompanyID() string {
	// Implementação simplificada - em produção usar UUID
	return "comp_" + time.Now().Format("20060102150405") + "_" + randomString(6)
}
