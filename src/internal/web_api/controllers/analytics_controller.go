package controllers

import (
	"encoding/json"
	"net/http"
	"shortener/src/internal/domain/visit"
	"shortener/src/pkg/logger"

	"github.com/go-chi/chi/v5"
)

type AnalyticsController struct {
	visitService visit.VisitService
}

func NewAnalyticsController(visitService visit.VisitService) *AnalyticsController {
	return &AnalyticsController{
		visitService: visitService,
	}
}

func (c *AnalyticsController) UseHandlers(r chi.Router) {
	r.Get("/analytics/{short_url}", c.Analytics)
}

// Analytics godoc
//
//	@Summary		Получить аналитику по короткой ссылке
//	@Description	Возвращает статистику переходов, агрегированную по дням, месяцам или User-Agent.
//	@Tags			analytics
//	@Param			short_url	path		string		true	"Короткий код"
//	@Param			group		query		string		true	"Тип группировки"	Enums(day,month,userAgent)
//	@Success		200			{object}	interface{}	"Результат зависит от типа группировки"
//	@Failure		400			{string}	string		"unknown group"
//	@Failure		500			{string}	string		"internal error"
//	@Router			/analytics/{short_url} [get]
func (c *AnalyticsController) Analytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shortURL := chi.URLParam(r, "short_url")
	group := r.URL.Query().Get("group")

	switch group {
	case "day":
		res, err := c.visitService.ByDayAnalytics(ctx, shortURL)
		if err != nil {
			logger.Error("failed to get analytics", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			logger.Error("failed to write response", "err", err)
		}

	case "month":
		res, err := c.visitService.ByMonthAnalytics(ctx, shortURL)
		if err != nil {
			logger.Error("failed to get analytics", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			logger.Error("failed to write response", "err", err)
		}

	case "userAgent":
		res, err := c.visitService.ByUserAgentAnalytics(ctx, shortURL)
		if err != nil {
			logger.Error("failed to get analytics", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			logger.Error("failed to write response", "err", err)
		}

	default:
		logger.Error("unknown group", "group", group)
		http.Error(w, "unknown group", http.StatusBadRequest)
	}
}
