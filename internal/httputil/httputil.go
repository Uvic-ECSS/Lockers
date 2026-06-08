package httputil

import (
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/parsa222/ECSS-Lockers/internal/crypto"
	"github.com/parsa222/ECSS-Lockers/internal/logger"
)

type ContextString string

const (
	SessionID ContextString = "session_id"
	Token     ContextString = "token"
)

var (
	assetVerOnce sync.Once
	assetVer     string
)

// changes only when assets/css/index.css changes, so ?v=<hash> browser caches per deploy
func assetVersion() string {
	assetVerOnce.Do(func() {
		b, err := os.ReadFile("assets/css/index.css")
		if err != nil {
			assetVer = "dev"
			return
		}
		sum := sha256.Sum256(b)
		assetVer = hex.EncodeToString(sum[:])[:10]
	})
	return assetVer
}

var tmplFuncs = template.FuncMap{
	"asset": func(p string) string { return p + "?v=" + assetVersion() },
}

// parse files with the shared func map
func NewTemplate(files ...string) (*template.Template, error) {
	return template.New(filepath.Base(files[0])).Funcs(tmplFuncs).ParseFiles(files...)
}

func WriteTemplateComponent(w http.ResponseWriter, data interface{}, filename ...string) {
	tmpl := template.Must(NewTemplate(filename...))

	w.WriteHeader(http.StatusOK)
	if err := tmpl.Execute(w, data); err != nil {
		logger.Error.Printf("error executing template data: %v\n", err)
		WriteResponse(w, http.StatusInternalServerError, nil)
	}
}

func WriteTemplatePage(w http.ResponseWriter, data interface{}, filename ...string) {
	files := make([]string, 1, len(filename)+1)

	files[0] = "templates/base.html"
	files = append(files, filename...)

	tmpl := template.Must(NewTemplate(files...))

	w.WriteHeader(http.StatusOK)
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		logger.Error.Printf("error executing template data: %v\n", err)
		WriteResponse(w, http.StatusInternalServerError, nil)
	}
}

func WriteResponse(w http.ResponseWriter, status int, writeData []byte) {
	w.WriteHeader(status)
	if writeData != nil {
		if _, err := w.Write(writeData); err != nil {
			logger.Error.Printf("failed to write response: %s\n", err)
		}
	}
}

func ExtractUserID(r *http.Request) string {
	sessionID, ok := r.Context().Value(SessionID).(string)
	if !ok {
		logger.Error.Fatal("ExtractUserID called from an unprotected route")
	}

	return sessionID
}

func ExtractUserEmail(r *http.Request) (string, error) {
	session := ExtractUserID(r)

	sessionID, err := crypto.Base64.DecodeString(session)
	if err != nil {
		return "", err
	}

	email, err := crypto.Decrypt(crypto.CipherKey[:], sessionID, nil)
	if err != nil {
		return "", err
	}

	return string(email), nil
}
