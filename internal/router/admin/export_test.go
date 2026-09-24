package admin

import (
	"bytes"
	"encoding/csv"
	"testing"
	"time"
)

func TestToCSV(t *testing.T) {
	expiry := time.Date(2026, 1, 16, 4, 30, 0, 0, time.UTC)
	out := toCSV([]locker_record{
		{LockerId: "ELW 001", UserName: "Doe, Jane", UserEmail: "jdoe@uvic.ca", ExpiryDate: expiry},
		{LockerId: "ELW 002", UserName: "=SUM(1,2)", UserEmail: "a@uvic.ca", ExpiryDate: expiry},
		{LockerId: "ELW 003", UserName: "José \"Pepe\"", UserEmail: "b@uvic.ca", ExpiryDate: expiry, ExpiryEmailSent: true},
	})

	if !bytes.HasPrefix(out, []byte("\ufeff")) {
		t.Fatal("missing UTF-8 marker for Excel")
	}
	rows, err := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(out, []byte("\ufeff")))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4 {
		t.Fatalf("got %d rows, want 4", len(rows))
	}

	want := [][]string{
		{"#", "Locker", "Name", "Email", "Expires", "Email Sent"},
		{"1", "ELW 001", "Doe, Jane", "jdoe@uvic.ca", "2026-01-15 20:30", "false"},
		{"2", "ELW 002", "'=SUM(1,2)", "a@uvic.ca", "2026-01-15 20:30", "false"},
		{"3", "ELW 003", "José \"Pepe\"", "b@uvic.ca", "2026-01-15 20:30", "true"},
	}
	for i := range want {
		for j := range want[i] {
			if rows[i][j] != want[i][j] {
				t.Errorf("row %d col %d = %q, want %q", i, j, rows[i][j], want[i][j])
			}
		}
	}
}
