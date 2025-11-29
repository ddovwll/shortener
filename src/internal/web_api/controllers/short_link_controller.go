package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	shortlink "shortener/src/internal/domain/short_link"
	"shortener/src/internal/domain/visit"
	"shortener/src/internal/web_api/models"
	"shortener/src/pkg/logger"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ShortLinkController struct {
	shortLinkService shortlink.ShortLinkService
	visitService     visit.VisitService
	validator        *validator.Validate
}

func NewShortLinkController(
	shortLinkService shortlink.ShortLinkService,
	visitService visit.VisitService,
	validator *validator.Validate,
) *ShortLinkController {
	return &ShortLinkController{
		shortLinkService: shortLinkService,
		visitService:     visitService,
		validator:        validator,
	}
}

func (c *ShortLinkController) UseHandlers(r chi.Router) {
	r.Post("/shorten", c.Create)
}

func (c *ShortLinkController) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req models.CreateShortLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.StructCtx(ctx, req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortLink, err := c.shortLinkService.Create(ctx, req.ShortURLString(), req.OriginalURL)
	if err != nil {
		if errors.Is(err, shortlink.ErrShortLinkAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		logger.Error("failed to create short link", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(shortLink); err != nil {
		logger.Error("failed to write response", err)
	}
}

func (c *ShortLinkController) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	shortURL := q.Get("short_url")

	shortLink, err := c.shortLinkService.Get(ctx, shortURL)
	if err != nil {
		if errors.Is(err, shortlink.ErrShortLinkNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		logger.Error("failed to get short link", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	visit := visit.Visit{
		ID:        uuid.New(),
		LinkID:    shortLink.ID,
		CreatedAt: time.Now(),
		UserAgent: r.UserAgent(),
		IPAddress: r.RemoteAddr,
	}

	if err := c.visitService.Register(ctx, visit); err != nil {
		logger.Error("failed to register visit", err)
	}

	http.Redirect(w, r, shortLink.OriginalURL, http.StatusFound)
}
