package admin

import (
	"context"
	"database/sql"
	"net/http"
	stdtime "time"

	"github.com/parsa222/ECSS-Lockers/internal/database"
	"github.com/parsa222/ECSS-Lockers/internal/httputil"
	"github.com/parsa222/ECSS-Lockers/internal/logger"
)

func Registrations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputil.WriteResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	if err := r.ParseForm(); err != nil {
		logger.Error.Println(err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	locker := r.FormValue("locker")

	db, lock := database.Lock()
	defer lock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*stdtime.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error.Println(err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO history (locker, user, name)
		SELECT locker, user, name FROM registration WHERE locker = :locker;`,
		sql.Named("locker", locker))
	if err != nil {
		logger.Error.Println(err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		logger.Error.Println(err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}
	if rowsAff == 0 {
		httputil.WriteResponse(w, http.StatusNotFound, nil)
		return
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM registration WHERE locker = :locker;`,
		sql.Named("locker", locker))
	if err != nil {
		logger.Error.Println(err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	if err = tx.Commit(); err != nil {
		logger.Error.Println(err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	logger.Trace.Printf("Deleted locker %s, result: %d row(s)\n", locker, rowsAff)
	httputil.WriteResponse(w, http.StatusNoContent, nil)
}

type historyEntry struct {
	Name    string
	Email   string
	Removed stdtime.Time
}

func History(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	locker := r.URL.Query().Get("locker")
	db, lock := database.Lock()
	defer lock.Unlock()

	rows, err := db.Query(`
		SELECT name, user, removed
		FROM history
		WHERE locker = :locker
		ORDER BY removed DESC;`, sql.Named("locker", locker))
	if err != nil {
		logger.Error.Println(err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}
	defer rows.Close()

	entries := make([]historyEntry, 0)
	for rows.Next() {
		entry := historyEntry{}
		if err := rows.Scan(&entry.Name, &entry.Email, &entry.Removed); err != nil {
			logger.Error.Println(err)
			httputil.WriteResponse(w, http.StatusInternalServerError, nil)
			return
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		logger.Error.Println(err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	httputil.WriteTemplateComponent(w, struct {
		Locker  string
		Entries []historyEntry
	}{locker, entries}, "templates/admin/historytable.html")
}
