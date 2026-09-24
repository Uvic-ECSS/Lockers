package admin

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Uvic-ECSS/Lockers/internal/httputil"
	"github.com/Uvic-ECSS/Lockers/internal/logger"
	"github.com/Uvic-ECSS/Lockers/internal/time"
)

func Export(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	lockers, err := queryAllRegistrations()
	if err != nil {
		logger.Error.Println(err)
		httputil.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	currentTerm := strings.ToLower(strings.ReplaceAll(formatTermName(time.GetCurrentTerm()), " ", ""))
	value := fmt.Sprintf("attachment; filename=registrations_%s.csv", currentTerm)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Add("Content-Disposition", value)
	httputil.WriteResponse(w, http.StatusOK, toCSV(lockers))
}

func toCSV(lockers []locker_record) []byte {
	var buf bytes.Buffer
	buf.WriteString("\ufeff")

	out := csv.NewWriter(&buf)
	out.Write([]string{"#", "Locker", "Name", "Email", "Expires", "Email Sent"})
	for i, locker := range lockers {
		out.Write([]string{
			strconv.Itoa(i + 1),
			safeCell(locker.LockerId),
			safeCell(locker.UserName),
			safeCell(locker.UserEmail),
			time.FormatSortable(locker.ExpiryDate),
			strconv.FormatBool(locker.ExpiryEmailSent),
		})
	}
	out.Flush()

	return buf.Bytes()
}

// Spreadsheets run a cell starting with one of these as a formula.
func safeCell(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}
