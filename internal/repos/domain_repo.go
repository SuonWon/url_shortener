package repo

import (
	"context"

	schema "github.com/url_shortener/internal/models"
	"gorm.io/gorm"
)

type DomainRepository interface {
	Create(context context.Context, domain *schema.CustomDomain) error
	GetDomains(context context.Context, page, pageSize int, q string) ([]GetDomainDTO, int64, error)
	GetDomainById(context context.Context, id int64) (*schema.CustomDomain, error)
	Update(context context.Context, domain *schema.CustomDomain) error
	Delete(context context.Context, id int64) error
}

type domainRepo struct {
	db *gorm.DB
}

type GetDomainDTO struct {
	ID         int64  `json:"id"`
	OwnerId    string `json:"owner_id"`
	OwnerName  string `json:"owner_name"`
	OwnerEmail string `json:"owner_email"`
	Domain     string `json:"domain"`
	IsVerified bool   `json:"is_verified"`
}

func NewDomainRepository(db *gorm.DB) DomainRepository {
	return &domainRepo{db: db}
}

// Create implements DomainRepository.
func (d *domainRepo) Create(context context.Context, domain *schema.CustomDomain) error {
	if err := d.db.WithContext(context).Create(domain).Error; err != nil {
		return err
	}
	return nil
}

// Delete implements DomainRepository.
func (d *domainRepo) Delete(context context.Context, id int64) error {
	if err := d.db.WithContext(context).Delete(&schema.CustomDomain{}, id).Error; err != nil {
		return err
	}
	return nil
}

// GetDomainById implements DomainRepository.
func (d *domainRepo) GetDomainById(context context.Context, id int64) (*schema.CustomDomain, error) {
	var domain schema.CustomDomain

	if err := d.db.WithContext(context).First(&domain, id).Error; err != nil {
		return nil, err
	}
	return &domain, nil
}

// GetDomains implements DomainRepository.
func (d *domainRepo) GetDomains(context context.Context, page int, pageSize int, q string) ([]GetDomainDTO, int64, error) {
	var (
		domains []GetDomainDTO
		count   int64
	)

	tx := d.db.WithContext(context).Table("custom_domains AS cd").Select("cd.id, cd.owner_id, users.name as owner_name, users.email as owner_email, cd.domain, cd.is_verified").Joins("LEFT JOIN users ON cd.owner_id = users.id")
	if q != "" {
		tx = tx.Where("owner ILIKE ? OR domain ILIKE ?", "%"+q+"%", "%"+q+"%")
	}
	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	if err := tx.Order("domain DESC").Limit(pageSize).Offset(offset).Find(&domains).Error; err != nil {
		return nil, 0, err
	}

	return domains, count, nil
}

// Update implements DomainRepository.
func (d *domainRepo) Update(context context.Context, domain *schema.CustomDomain) error {
	return d.db.WithContext(context).Save(domain).Error
}
