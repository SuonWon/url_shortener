package user_handler

import (
	"context"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	schema "github.com/url_shortener/internal/models"
	repo "github.com/url_shortener/internal/repos"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	repo      repo.UserRepository
	validator *validator.Validate
}

type GetUserDTO struct {
	Id    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
}

type CreateUserDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=7"`
	Name     string `json:"name" validate:"required,min=2"`
}

type UpdateUserDTO struct {
	ID    string  `json:"id" validate:"required"`
	Email *string `json:"email" validate:"required,email"`
	Name  *string `json:"name" validate:"required,min=2"`
}

func NewUserHandler(repo repo.UserRepository) *UserHandler {
	return &UserHandler{
		repo:      repo,
		validator: validator.New(),
	}
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

func Map[T any, R any](in []T, f func(T) R) []R {
	out := make([]R, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

func (u *UserHandler) Create(context *gin.Context) {
	var body CreateUserDTO
	if err := context.ShouldBindJSON(&body); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "invalid json: " + err.Error()})
		return
	}

	if err := u.validator.Struct(body); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx, cancel := withTimeout(context)
	defer cancel()

	newUUID, _ := exec.Command("uuidgen").Output()
	hash, _ := HashPasswordBcrypt(body.Password)
	user := schema.User{
		Email:    body.Email,
		Name:     body.Name,
		Password: hash,
		Salt:     strings.TrimSpace(string(newUUID)),
	}

	if err := u.repo.Create(ctx, &user); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "create failed: " + err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "new user added!"})
}

func (u *UserHandler) GetUserById(context *gin.Context) {
	id := context.Param("id")
	if id == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx, cancel := withTimeout(context)
	defer cancel()

	res, err := u.repo.GetUserById(ctx, id)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	}

	user := GetUserDTO{
		Id:    res.Id,
		Name:  res.Name,
		Email: res.Email,
	}

	context.JSON(http.StatusOK, gin.H{"data": user})
}

func (u *UserHandler) GetUsers(context *gin.Context) {
	page := atoiDefault(context.Query("page"), 1)
	size := atoiDefault(context.Query("pageSize"), 20)
	q := context.Query("q")

	ctx, cancel := withTimeout(context)
	defer cancel()

	res, total, err := u.repo.GetUsers(ctx, page, size, q)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "list failed: " + err.Error()})
		return
	}

	users := Map(res, func(item schema.User) GetUserDTO {
		return GetUserDTO{Id: item.Id, Email: item.Email, Name: item.Name}
	})

	context.JSON(http.StatusOK, gin.H{
		"data":       users,
		"total":      total,
		"page":       page,
		"pageSize":   size,
		"totalPages": (total + int64(size) - 1) / int64(size),
	})
}

func (u *UserHandler) DeleteUser(context *gin.Context) {
	id := context.Param("id")
	if id == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx, cancel := withTimeout(context)
	defer cancel()

	if err := u.repo.Delete(ctx, id); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "create failed: " + err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User deleted successfully!"})
}

func (u *UserHandler) UpdateUser(context *gin.Context) {
	var body UpdateUserDTO
	if err := context.ShouldBindJSON(&body); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "invalid json: " + err.Error()})
		return
	}

	if err := u.validator.Struct(body); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx, cancel := withTimeout(context)
	defer cancel()

	user, err := u.repo.GetUserById(ctx, body.ID)
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if body.Name != nil {
		user.Name = *body.Name
	}

	if body.Email != nil {
		user.Email = *body.Email
	}

	if err := u.repo.Update(ctx, user); err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"error": "error in updating user"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "user updated successfully"})
}

func withTimeout(c *gin.Context) (ctx context.Context, cancel func()) {
	ctx1, cancel := c.Request.Context(), func() {}
	return ctx1, cancel
}

const bcryptCost = 12

func HashPasswordBcrypt(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(hash), err // store this string
}

func CheckPasswordBcrypt(storedHash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) == nil
}
