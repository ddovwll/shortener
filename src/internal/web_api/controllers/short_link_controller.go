package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"shortener/src/internal/application/services"
	shortlink "shortener/src/internal/domain/short_link"
	"shortener/src/internal/domain/visit"
	"shortener/src/internal/web_api/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type ShortLinkController struct {
	shortLinkService services.ShortLinkService
	visitService     visit.VisitService
	validator        *validator.Validate
}

func NewShortLinkController(
	shortLinkService services.ShortLinkService,
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
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(shortLink); err != nil {
		// todo log
	}
}

func (c *ShortLinkController) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	shortUrl := q.Get("short_url")

	shortLink, err := c.shortLinkService.Get(ctx, shortUrl)
	if err != nil {
		if errors.Is(err, shortlink.ErrShortLinkNotFound) {
			// todo Not Found page
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}

	http.Redirect(w, r, shortLink.OriginalURL, http.StatusFound)
}
