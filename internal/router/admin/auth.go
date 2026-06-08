package admin

import (
	"crypto/subtle"
	"encoding/binary"
	"net"
	"net/http"
	"sync"

	"github.com/parsa222/ECSS-Lockers/internal/crypto"
	"github.com/parsa222/ECSS-Lockers/internal/env"
	"github.com/parsa222/ECSS-Lockers/internal/httputil"
	"github.com/parsa222/ECSS-Lockers/internal/logger"
	"github.com/parsa222/ECSS-Lockers/internal/time"
)

const (
	adminSessionMaxAge int    = 8 * 60 * 60 // 8 h
	loginWindowSecs    uint64 = 15 * 60
	loginMaxFails      int    = 5 // IP limit
)

var (
	adminUsername string
	adminPassword string

	failsMu sync.Mutex
	fails   = map[string][]uint64{} // unix time
)

func Initialize() {
	adminUsername = env.MustEnv("ADMIN_USERNAME")
	adminPassword = env.MustEnv("ADMIN_PASSWORD")
}

func Auth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodGet {
		httputil.WriteResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	if r.Method == http.MethodGet {
		httputil.WriteTemplatePage(w, nil, "templates/admin/auth.html")
		return
	}

	ip := clientIP(r)

	if tooManyFailures(ip) {
		httputil.WriteResponse(w, http.StatusOK, []byte(`
    <button type="submit" class="btn btn-primary btn-block">Login</button>
    <span id="error-message" class="text-error">Too many attempts : ( </span>
            `))
		return
	}

	if err := r.ParseForm(); err != nil {
		logger.Error.Println(err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	userOK := subtle.ConstantTimeCompare([]byte(username), []byte(adminUsername))
	passOK := subtle.ConstantTimeCompare([]byte(password), []byte(adminPassword))
	if userOK&passOK != 1 {
		recordFailure(ip)
		httputil.WriteResponse(w, http.StatusOK, []byte(`
    <button type="submit" class="btn btn-primary btn-block">Login</button>
    <span id="error-message" class="text-error">Invalid credentials.</span>
            `))
		return
	}

	clearFailures(ip)

	token, err := makeToken()
	if err != nil {
		logger.Error.Println(err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    token,
		Path:     "/",
		MaxAge:   adminSessionMaxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Add("HX-Redirect", "/admin")
}

func AdminTokenChecker(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("admin_token")
		if err != nil || !validToken(cookie.Value) {
			httputil.WriteTemplatePage(w,
				struct{ IsAdmin bool }{IsAdmin: true},
				"templates/base.html",
				"templates/auth/session_expired.html",
				"templates/nav.html")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// server side token expiry date
func makeToken() (string, error) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(time.Now().Unix()))

	ciphertext, err := crypto.Encrypt(crypto.CipherKey[:], buf, []byte(adminUsername))
	if err != nil {
		return "", err
	}

	return crypto.Base64.EncodeToString(ciphertext), nil
}

func validToken(token string) bool {
	raw, err := crypto.Base64.DecodeString(token)
	if err != nil {
		return false
	}

	pt, err := crypto.Decrypt(crypto.CipherKey[:], raw, []byte(adminUsername))
	if err != nil || len(pt) < 8 {
		return false
	}

	issued := binary.BigEndian.Uint64(pt[:8])
	now := uint64(time.Now().Unix())
	return now >= issued && now-issued <= uint64(adminSessionMaxAge)
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func tooManyFailures(ip string) bool {
	failsMu.Lock()
	defer failsMu.Unlock()

	now := uint64(time.Now().Unix())
	var kept []uint64
	for _, t := range fails[ip] {
		if now-t < loginWindowSecs {
			kept = append(kept, t)
		}
	}

	if len(kept) == 0 {
		delete(fails, ip)
	} else {
		fails[ip] = kept
	}

	return len(kept) >= loginMaxFails
}

func recordFailure(ip string) {
	failsMu.Lock()
	defer failsMu.Unlock()
	fails[ip] = append(fails[ip], uint64(time.Now().Unix()))
}

func clearFailures(ip string) {
	failsMu.Lock()
	defer failsMu.Unlock()
	delete(fails, ip)
}
