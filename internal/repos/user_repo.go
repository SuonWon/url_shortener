package user_repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	schema "github.com/url_shortener/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(context context.Context, user *schema.User) error
	GetUsers(context context.Context, page, pageSize int, q string) ([]schema.User, int64, error)
	GetUserById(context context.Context, id string) (*schema.User, error)
	Update(context context.Context, user *schema.User) error
	Delete(context context.Context, id string) error
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

// Create implements UserRepository.
func (u *userRepo) Create(context context.Context, user *schema.User) error {
	if err := u.db.WithContext(context).Create(user).Error; err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

// Delete implements UserRepository.
func (u *userRepo) Delete(context context.Context, id string) error {

	userId, err := uuid.Parse(id)
	if err != nil {
		// not a valid UUID
		panic(err)
	}

	return u.db.WithContext(context).Delete(&schema.User{}, userId).Error
}

// GetUserById implements UserRepository.
func (u *userRepo) GetUserById(context context.Context, id string) (*schema.User, error) {
	var user schema.User

	userId, err := uuid.Parse(id)
	if err != nil {
		// not a valid UUID
		panic(err)
	}

	if err := u.db.WithContext(context).First(&user, userId).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUsers implements UserRepository.
func (u *userRepo) GetUsers(context context.Context, page int, pageSize int, q string) ([]schema.User, int64, error) {
	var (
		users []schema.User
		count int64
	)
	tx := u.db.WithContext(context).Model(&schema.User{})
	if q != "" {
		tx = tx.Where("name ILIKE ? OR email ILIKE ?", "%"+q+"%", "%"+q+"%")
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

	if err := tx.Order("name DESC").Limit(pageSize).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, count, nil
}

// Update implements UserRepository.
func (u *userRepo) Update(context context.Context, user *schema.User) error {
	return u.db.WithContext(context).Save(user).Error
}
