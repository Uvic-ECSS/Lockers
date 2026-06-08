package dash

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	stdtime "time"

	"github.com/parsa222/ECSS-Lockers/internal/database"
	"github.com/parsa222/ECSS-Lockers/internal/httputil"
	"github.com/parsa222/ECSS-Lockers/internal/logger"
	"github.com/parsa222/ECSS-Lockers/internal/time"
)

type lockerState struct {
	IsAvailable bool
	LockerID    string
}

type dashboardData struct {
	HasLocker  bool
	LockerName string
	ExpireAt   string
	IsExpired  bool
}

func Dash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	userEmail, err := httputil.ExtractUserEmail(r)
	if err != nil {
		logger.Error.Printf("failed to extract user email from token: %v\n", err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	data, err := userDashboardData(userEmail)
	if err != nil {
		logger.Error.Printf("error querying for registration: %v\n", err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	httputil.WriteTemplatePage(w, data,
		"templates/nav.html",
		"templates/dash/index.html",
		"templates/dash/locker_status.html")
}

func userDashboardData(userEmail string) (dashboardData, error) {
	var data dashboardData

	db, lock := database.Lock()
	defer lock.Unlock()

	stmt, err := db.Prepare(`
        SELECT locker, expiry
        FROM registration
        WHERE user = :email
        LIMIT 1;`)
	if err != nil {
		return data, err
	}
	defer stmt.Close()

	var expiry stdtime.Time
	err = stmt.QueryRow(sql.Named("email", userEmail)).Scan(&data.LockerName, &expiry)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return data, nil
		}
		return data, err
	}

	data.HasLocker = true
	data.ExpireAt = expiry.Format(time.TimeFormatLayout)
	data.IsExpired = expiry.Before(time.Now())
	return data, nil
}

func ApiLocker(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		logger.Error.Printf("failed to parse form: %v\n", err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	locker := r.FormValue("locker")
	if len(locker) == 0 {
		httputil.WriteResponse(w, http.StatusOK, nil)
		return
	}

	lockerNum, err := strconv.ParseUint(locker, 10, 16)
	if err != nil {
		httputil.WriteResponse(
			w,
			http.StatusOK,
			[]byte("<p class=\"text-error text-center\">Invalid locker</p>"))
		return
	}

	db, lock := database.Lock()
	defer lock.Unlock()

	stmt, err := db.Prepare(`
        SELECT locker.id, registration.locker 
        FROM locker
        LEFT JOIN registration 
        ON locker.id = registration.locker
        WHERE locker.id 
        LIKE ?;`)

	if err != nil {
		logger.Error.Fatal("stmt error:", err)
	}

	//locker = fmt.Sprintf("%%ELW %d%%", lockerNum)
	locker = fmt.Sprintf("%%%d%%", lockerNum)
	// this should fix the bug and make it easier to add the ecs building later

	rows, err := stmt.Query(locker)
	if err != nil {
		panic(err)
	}

	lockers := []lockerState{}
	for rows.Next() {
		var (
			lockerID       string
			registrationID sql.NullString
		)

		if err := rows.Scan(&lockerID, &registrationID); err != nil {
			logger.Error.Printf("failed to scan data: %v\n", err)
			httputil.WriteResponse(w, http.StatusInternalServerError, nil)
			return
		}

		lockers = append(lockers, lockerState{
			IsAvailable: !registrationID.Valid,
			LockerID:    lockerID,
		})
	}

	data := struct {
		LockerOK bool
		Lockers  []lockerState
	}{
		LockerOK: len(lockers) != 0,
		Lockers:  lockers,
	}

	httputil.WriteTemplateComponent(w, data, "templates/dash/locker_card.html")
}

func DashLockerRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodGet {
		httputil.WriteResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	if err := r.ParseForm(); err != nil {
		logger.Error.Printf("error parsing form: %v\n", err)
		httputil.WriteResponse(w, http.StatusBadRequest, nil)
		return
	}

	locker := r.FormValue("locker")

	if r.Method == http.MethodGet {
		httputil.WriteTemplatePage(w, locker,
			"templates/nav.html", "templates/dash/locker_register.html")
		return
	}

	userName := r.FormValue("name")

	userEmail, err := httputil.ExtractUserEmail(r)
	if err != nil {
		logger.Error.Printf("error decrypting user email: %v\n", err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	db, lock := database.Lock()
	defer lock.Unlock()

	var stmt *sql.Stmt

	stmt, err = db.Prepare(`
        SELECT COUNT(*) 
        FROM registration 
        WHERE locker = :locker;`)

	if err != nil {
		logger.Error.Fatal(err)
	}

	var registrationCount uint8

	err = stmt.QueryRow(sql.Named("locker", locker)).Scan(&registrationCount)
	if err != nil {
		logger.Error.Printf("error querying for locker: %v\n", err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	if registrationCount != 0 {
		httputil.WriteTemplateComponent(w, nil, "templates/dash/locker_unavailable.html")
		return
	}

	stmt, err = db.Prepare(`
        INSERT INTO registration (locker, user, name, expiry)
        VALUES (:locker, :user, :name, :expiry);`)

	if err != nil {
		logger.Error.Fatal(err)
	}

	expiryDate := time.NextExpiryDate(time.Now())

	_, err = stmt.Exec(
		sql.Named("locker", locker),
		sql.Named("user", userEmail),
		sql.Named("name", userName),
		sql.Named("expiry", expiryDate))

	if err != nil {
		logger.Error.Printf("error writing registration to db: %v\n", err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	httputil.WriteTemplateComponent(w, nil, "templates/dash/locker_register_ok.html")

}
func DashDeregister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	userEmail, err := httputil.ExtractUserEmail(r)
	if err != nil {
		logger.Error.Printf("error extracting user email: %v\n", err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	db, lock := database.Lock()
	defer lock.Unlock()

	stmt, err := db.Prepare(`DELETE FROM registration WHERE user = :email`)
	if err != nil {
		logger.Error.Printf("failed to prepare delete statement: %v\n", err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(sql.Named("email", userEmail))
	if err != nil {
		logger.Error.Printf("error deregistering locker: %v\n", err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	httputil.WriteTemplateComponent(w, nil, "templates/dash/locker_register_ok.html")

}

func DashRenew(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	userEmail, err := httputil.ExtractUserEmail(r)
	if err != nil {
		logger.Error.Printf("error extracting user email: %v\n", err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	if err := renewRegistration(userEmail); err != nil {
		if errors.Is(err, errNotExpired) {
			httputil.WriteResponse(w, http.StatusConflict, nil)
			return
		}
		logger.Error.Printf("error renewing locker: %v\n", err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	data, err := userDashboardData(userEmail)
	if err != nil {
		logger.Error.Printf("error querying for registration: %v\n", err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	httputil.WriteTemplateComponent(w, data, "templates/dash/locker_status.html")
}

var errNotExpired = errors.New("locker is not expired yet ;)")

func renewRegistration(userEmail string) error {
	db, lock := database.Lock()
	defer lock.Unlock()

	sel, err := db.Prepare(`SELECT expiry FROM registration WHERE user = :email;`)
	if err != nil {
		return err
	}
	var expiry stdtime.Time
	err = sel.QueryRow(sql.Named("email", userEmail)).Scan(&expiry)
	sel.Close()
	if err != nil {
		return err
	}
	if !expiry.Before(time.Now()) {
		return errNotExpired
	}

	stmt, err := db.Prepare(`
        UPDATE registration
        SET expiry = :expiry, expiryEmailSent = FALSE
        WHERE user = :email;`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		sql.Named("expiry", time.NextExpiryDate(time.Now())),
		sql.Named("email", userEmail))
	return err
}
