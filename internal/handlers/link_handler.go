package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	schema "github.com/url_shortener/internal/models"
	repo "github.com/url_shortener/internal/repos"
	tiny "github.com/url_shortener/internal/util"
)

type LinkHandler struct {
	repo       repo.LinkRepository
	domainRepo repo.DomainRepository
	validator  *validator.Validate
}

func NewLinkHandler(repo repo.LinkRepository, domainRepo repo.DomainRepository) *LinkHandler {
	return &LinkHandler{
		repo:       repo,
		domainRepo: domainRepo,
		validator:  validator.New(),
	}
}

type CreateShortLinkDTO struct {
	OwnerId   string `json:"owner_id" validate:"required"`
	DomainId  uint64 `json:"domain_id" validate:"required"`
	TargetURL string `json:"target_url" validate:"required"`
}

func (l *LinkHandler) Create(context *gin.Context) {
	var body CreateShortLinkDTO
	if err := context.ShouldBind(&body); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "invalid json: " + err.Error()})
		return
	}

	if err := l.validator.Struct(body); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx, cancel := withTimeout(context)
	defer cancel()

	// domain := l.domainRepo.GetDomainById(context, )
	code, err := tiny.RandCode(7)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "error in generate code"})
		return
	}
	ownerId, _ := uuid.Parse(body.OwnerId)
	shortLink := schema.ShortLink{
		OwnerId:   ownerId,
		DomainId:  &body.DomainId,
		Code:      code,
		TargetURL: body.TargetURL,
		IsActive:  true,
		ExpiredAt: &time.Time{},
	}

	if err := l.repo.Create(ctx, &shortLink); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "create filed!"})
		return
	}

	domain, err := l.domainRepo.GetDomainById(context, int64(body.DomainId))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "domain not found"})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"short_link": domain.Domain + "/" + code})
}

func (l LinkHandler) RedirectLink(context *gin.Context) {
	code := context.Param("code")
	if code == "" {
		context.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}

	ctx, cancel := withTimeout(context)
	defer cancel()

	res, err := l.repo.GetLinkByCode(ctx, code)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "url not found!"})
		return
	}
	http.Redirect(context.Writer, context.Request, res, http.StatusMovedPermanently)

}

func (l LinkHandler) Test(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"message": "Reidrected"})
}
