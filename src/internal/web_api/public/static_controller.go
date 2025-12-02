package public

import (
	"embed"
	"net/http"
	"shortener/src/pkg/logger"

	"github.com/go-chi/chi/v5"
)

//go:embed *.html
var htmlFS embed.FS

func UseStaticFiles(r chi.Router) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := htmlFS.ReadFile("index.html")
		if err != nil {
			http.Error(w, "index.html not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, err = w.Write(data)
		if err != nil {
			logger.Error("failed to write index.html")
		}
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		data, err := htmlFS.ReadFile("not_found.html")
		if err != nil {
			http.Error(w, "not_found.html not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_, err = w.Write(data)
		if err != nil {
			logger.Error("failed to write not_found.html")

		}
	})
}
