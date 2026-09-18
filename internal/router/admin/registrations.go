package admin

import (
	"context"
	"database/sql"
	"net/http"
	stdtime "time"

	"github.com/Uvic-ECSS/ECSS-Lockers/internal/database"
	"github.com/Uvic-ECSS/ECSS-Lockers/internal/httputil"
	"github.com/Uvic-ECSS/ECSS-Lockers/internal/logger"
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
		INSERT INTO locker_removals (locker_id, user_email, user_name)
		SELECT locker_id, user_email, user_name
		FROM locker_registrations WHERE locker_id = :locker;`,
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

	_, err = tx.ExecContext(ctx, `DELETE FROM locker_registrations WHERE locker_id = :locker;`,
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
	UserName  string
	UserEmail string
	Removed   string
}

func History(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	locker := r.URL.Query().Get("locker")
	db, lock := database.Lock()

	rows, err := db.Query(`
		SELECT user_name, user_email, removed_date
		FROM locker_removals
		WHERE locker_id = :locker
		ORDER BY removed_date DESC;`, sql.Named("locker", locker))
	lock.Unlock()
	if err != nil {
		logger.Error.Println(err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}
	defer rows.Close()

	entries := make([]historyEntry, 0)
	for rows.Next() {
		entry := historyEntry{}
		var removed stdtime.Time
		if err := rows.Scan(&entry.UserName, &entry.UserEmail, &removed); err != nil {
			logger.Error.Println(err)
			httputil.WriteResponse(w, http.StatusInternalServerError, nil)
			return
		}
		entry.Removed = removed.Format("Jan 2, 2006 at 3:04pm")
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
	}{locker, entries}, "templates/admin/history_table.html")
}
