package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	schema "github.com/url_shortener/internal/models"
	repo "github.com/url_shortener/internal/repos"
)

type DomainHandler struct {
	repo      repo.DomainRepository
	validator *validator.Validate
}

type CreateDomainDTO struct {
	OwnerId string `json:"owner_id" validate:"required"`
	Domain  string `json:"domain" validate:"required"`
}

type GetDoaminDTO struct {
	ID         int64  `json:"id"`
	OwnerId    string `json:"owner_id"`
	Domain     string `json:"domain"`
	IsVerified bool   `json:"is_verified"`
}

func NewDomainRepository(repo repo.DomainRepository) *DomainHandler {
	return &DomainHandler{
		repo:      repo,
		validator: validator.New(),
	}
}

func (d *DomainHandler) Create(context *gin.Context) {
	var body CreateDomainDTO
	if err := context.ShouldBind(&body); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "invalid join: " + err.Error()})
		return
	}

	if err := d.validator.Struct(body); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx, cancel := withTimeout(context)
	defer cancel()

	domain := schema.CustomDomain{
		OwnerId: body.OwnerId,
		Domain:  body.Domain,
	}

	if err := d.repo.Create(ctx, &domain); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "create failed: " + err.Error()})
	}

	context.JSON(http.StatusCreated, gin.H{"message": "new domain added!"})
}

func (d *DomainHandler) GetDomains(context *gin.Context) {
	page := atoiDefault(context.Query("page"), 1)
	size := atoiDefault(context.Query("pageSize"), 20)
	q := context.Query("q")

	ctx, cancel := withTimeout(context)
	defer cancel()

	res, total, err := d.repo.GetDomains(ctx, page, size, q)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "list failed: " + err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"data":       res,
		"total":      total,
		"page":       page,
		"pageSize":   size,
		"totalPages": (total + int64(size) - 1) / int64(size),
	})
}

func (d *DomainHandler) GetDomainById(context *gin.Context) {
	id := context.Param("id")
	if id == "" {
		context.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}

	ctx, cancel := withTimeout(context)
	defer cancel()

	domainId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}

	res, err := d.repo.GetDomainById(ctx, domainId)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "doamin not found"})
		return
	}

	domain := GetDoaminDTO{
		ID:         res.Id,
		OwnerId:    res.OwnerId,
		Domain:     res.Domain,
		IsVerified: res.IsVerified,
	}

	context.JSON(http.StatusOK, gin.H{"data": domain})
}
