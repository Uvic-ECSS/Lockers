package admin

import (
	"database/sql"
	"net/http"
	"strconv"
	stdtime "time"

	"github.com/parsa222/ECSS-Lockers/internal/database"
	"github.com/parsa222/ECSS-Lockers/internal/httputil"
	"github.com/parsa222/ECSS-Lockers/internal/logger"
	"github.com/parsa222/ECSS-Lockers/internal/time"
)

type locker_record struct {
	RowIndex        uint16
	LockerId        string
	UserName        string
	UserEmail       string
	ExpiryDate      stdtime.Time
	ExpiryEmailSent bool
	Expiry          string
	RemovedDate     stdtime.Time
	Removed         string
}

func Home(w http.ResponseWriter, r *http.Request) {
	data := struct {
		AllLockers             []locker_record
		HasLockers             bool
		RegisteredLockers      []locker_record
		HasRegistrations       bool
		UnregisteredLockers    []locker_record
		HasUnregisteredLockers bool
		Term                   string
	}{
		Term: formatTermName(stdtime.Now()),
	}

	var err error

	data.AllLockers, err = queryAllLockers()
	if err != nil {
		logger.Error.Println(err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}
	if len(data.AllLockers) > 0 {
		data.HasLockers = true
	}

	data.RegisteredLockers, err = queryAllRegistrations()
	if err != nil {
		logger.Error.Println(err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}
	if len(data.RegisteredLockers) > 0 {
		data.HasRegistrations = true
	}

	data.UnregisteredLockers = FilterUnregistered(data.AllLockers, data.RegisteredLockers)
	if len(data.UnregisteredLockers) > 0 {
		data.HasUnregisteredLockers = true
	}

	httputil.WriteTemplatePage(
		w,
		data,
		"templates/nav.html",
		"templates/admin/index.html",
		"templates/admin/registered_table.html",
		"templates/admin/unregistered_table.html",
		"templates/admin/history_table.html")
}

func queryAll[T any](query string, scan func(*sql.Rows, uint16) (T, error)) ([]T, error) {
	db, lock := database.Lock()

	rows, err := db.Query(query)
	lock.Unlock()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]T, 0, 200)
	rowIndex := uint16(1)

	for rows.Next() {
		item, err := scan(rows, rowIndex)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
		rowIndex++
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func queryAllRegistrations() ([]locker_record, error) {
	return queryAll(`
		SELECT locker_id, user_email, user_name, expiry_date, expiry_email_sent
		FROM locker_registrations
		ORDER BY locker_id;`,
		func(rows *sql.Rows, rowIndex uint16) (locker_record, error) {
			reg := locker_record{RowIndex: rowIndex}
			err := rows.Scan(
				&reg.LockerId,
				&reg.UserEmail,
				&reg.UserName,
				&reg.ExpiryDate,
				&reg.ExpiryEmailSent,
			)
			if err != nil {
				return reg, err
			}
			reg.Expiry = reg.ExpiryDate.Format(time.TimeFormatLayout)
			return reg, nil
		},
	)
}

func queryAllLockers() ([]locker_record, error) {
	return queryAll(`
		SELECT id
		FROM locker
		ORDER BY id;`,
		func(rows *sql.Rows, rowIndex uint16) (locker_record, error) {
			l := locker_record{RowIndex: rowIndex}
			err := rows.Scan(&l.LockerId)
			return l, err
		},
	)
}

func queryAllLockerRemovals() ([]locker_record, error) {
	return queryAll(`
		SELECT locker_id, user_email, user_name, removed_date
		FROM locker_removals
		ORDER BY removed_date DESC;`,
		func(rows *sql.Rows, rowIndex uint16) (locker_record, error) {
			lr := locker_record{RowIndex: rowIndex}
			err := rows.Scan(&lr.LockerId, &lr.UserEmail, &lr.UserName, &lr.RemovedDate)
			if err != nil {
				return lr, err
			}
			lr.Removed = lr.RemovedDate.Format(time.TimeFormatLayout)
			return lr, nil
		},
	)
}

func formatTermName(t stdtime.Time) string {
	year := t.Year()
	month := t.Month()
	yearStr := strconv.Itoa(year)

	switch {
	case month >= stdtime.September && month <= stdtime.December:
		return "Fall " + yearStr
	case month >= stdtime.January && month <= stdtime.April:
		return "Spring " + yearStr
	default:
		return "Summer " + yearStr
	}
}

// FilterUnregistered returns lockers present in allLockers but absent from registeredLockers
func FilterUnregistered(allLockers, registeredLockers []locker_record) []locker_record {
	excludeMap := make(map[string]struct{}, len(registeredLockers))

	for _, reg := range registeredLockers {
		excludeMap[normalizeLockerId(reg.LockerId)] = struct{}{}
	}

	unregistered := make([]locker_record, 0, len(allLockers))
	var rowIndex uint16 = 1

	for _, locker := range allLockers {
		if _, excluded := excludeMap[normalizeLockerId(locker.LockerId)]; !excluded {
			item := locker
			item.RowIndex = rowIndex
			unregistered = append(unregistered, item)
			rowIndex++
		}
	}

	return unregistered
}

func normalizeLockerId(lockerId string) string {
	return lockerId
}
