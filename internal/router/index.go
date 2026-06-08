package router

import (
	"html/template"
	"net/http"

	"github.com/parsa222/ECSS-Lockers/internal"
	"github.com/parsa222/ECSS-Lockers/internal/httputil"
	"github.com/parsa222/ECSS-Lockers/internal/logger"
)

func Home(w http.ResponseWriter, r *http.Request) {
	// Cache for 15 mins
	w.Header().Add("Cache-Control", "max-age=900")
	w.WriteHeader(http.StatusOK)

	// Parse the template files
	httputil.WriteTemplatePage(w, struct{ Debug bool }{Debug: internal.Debug}, "templates/index.html")
}

func SessionExpired(w http.ResponseWriter, r *http.Request) {
	// Cache for 15 mins
	w.Header().Add("Cache-Control", "max-age=900")
	w.WriteHeader(http.StatusOK)

	// Parse the template files
	httputil.WriteTemplatePage(w, nil, "templates/auth/session_expired.html", "templates/nav.html")
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/base.html", "templates/404.html")
	if err != nil {
		logger.Error.Printf("error parsing 404 page: %v\n", err)
		httputil.WriteResponse(w, http.StatusNotFound, []byte("404 - not found"))
		return
	}

	w.WriteHeader(http.StatusNotFound)
	if err := tmpl.ExecuteTemplate(w, "base", nil); err != nil {
		logger.Error.Printf("error rendering 404 page: %v\n", err)
	}
}
