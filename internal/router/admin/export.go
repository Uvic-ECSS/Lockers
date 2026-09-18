package admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Uvic-ECSS/ECSS-Lockers/internal/httputil"
	"github.com/Uvic-ECSS/ECSS-Lockers/internal/logger"
	"github.com/Uvic-ECSS/ECSS-Lockers/internal/time"
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
	w.Header().Add("Content-Disposition", value)
	httputil.WriteResponse(w, http.StatusOK, toCSV(lockers))
}

func toCSV(lockers []locker_record) []byte {
	buf := make([]string, len(lockers)+1)
	buf[0] = ",Locker,Name,Email,Expire On, Email Sent"

	for i, locker := range lockers {
		sent := "false"
		if locker.ExpiryEmailSent {
			sent = "true"
		}
		buf[i+1] = fmt.Sprintf(
			"%d,%s,%s,%s,%s,%s",
			i+1, locker.LockerId,
			locker.UserName, locker.UserEmail,
			locker.ExpiryDate.Format("2006-01-02 15:04:05 MST"),
			sent)
	}

	return []byte(strings.Join(buf, "\n"))
}
