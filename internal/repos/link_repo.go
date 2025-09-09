package repo

import (
	"context"

	schema "github.com/url_shortener/internal/models"
	"gorm.io/gorm"
)

type LinkRepository interface {
	Create(context context.Context, link *schema.ShortLink) error
	GetLinks(context context.Context, page, pageSize int) ([]schema.ShortLink, int64, error)
	GetLinkByCode(context context.Context, code string) (string, error)
	Delete(context context.Context, id int64) error
}

type linkRepo struct {
	db *gorm.DB
}

func NewShortLinkRepository(db *gorm.DB) LinkRepository {
	return &linkRepo{db: db}
}

// Create implements LinkRepository.
func (l *linkRepo) Create(context context.Context, link *schema.ShortLink) error {
	if err := l.db.WithContext(context).Create(link).Error; err != nil {
		return err
	}
	return nil
}

// Delete implements LinkRepository.
func (l *linkRepo) Delete(context context.Context, id int64) error {
	return l.db.WithContext(context).Delete(&schema.ShortLink{}, id).Error
}

// GetLinkByCode implements LinkRepository.
func (l *linkRepo) GetLinkByCode(context context.Context, code string) (string, error) {
	var link schema.ShortLink

	if err := l.db.WithContext(context).Where("code = ?", code).Find(&link).Error; err != nil {
		return "", err
	}
	return link.TargetURL, nil
}

// GetLinks implements LinkRepository.
func (l *linkRepo) GetLinks(context context.Context, page int, pageSize int) ([]schema.ShortLink, int64, error) {
	var (
		links []schema.ShortLink
		count int64
	)

	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	if err := l.db.WithContext(context).Order("id DESC").Limit(pageSize).Offset(offset).Error; err != nil {
		return nil, 0, err
	}

	return links, count, nil
}
