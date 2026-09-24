package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/Uvic-ECSS/Lockers/internal/database"
	"github.com/Uvic-ECSS/Lockers/internal/env"
	"github.com/Uvic-ECSS/Lockers/internal/logger"
	"github.com/joho/godotenv"
)

var moves = []struct{ table, copy, backup string }{
	{
		"locker",
		`INSERT OR IGNORE INTO lockers (locker_id)
		SELECT id FROM locker WHERE id IS NOT NULL AND id <> '';`,
		`ALTER TABLE locker RENAME TO locker_backup;`,
	},
	{
		"registration",
		`INSERT OR IGNORE INTO locker_registrations
			(locker_id, user_email, user_name, expiry_date, expiry_email_sent)
		SELECT locker, user, name, expiry, expiryEmailSent FROM registration;`,
		`ALTER TABLE registration RENAME TO registration_backup;`,
	},
}

func main() {
	if err := godotenv.Load(); err != nil {
		logger.Error.Fatal(err)
	}

	logger.Info.Println("DATABASE MIGRATION")

	database.Connect(fmt.Sprintf(
		"%s?authToken=%s",
		env.MustEnv("DATABASE_URL"),
		env.MustEnv("DATABASE_AUTH_TOKEN")))

	db, lock := database.Lock()
	defer lock.Unlock()

	if err := migrate(db); err != nil {
		logger.Error.Fatal(err)
	}

	if err := seed(db); err != nil {
		logger.Error.Fatal(err)
	}

	logger.Info.Println("Done")
}

// all or nothing....
func migrate(db *sql.DB) error {
	schema, err := os.ReadFile("internal/database/schema.sql")
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(schema)); err != nil {
		return err
	}
	logger.Info.Println("created schema.")

	for _, m := range moves {
		var exists bool
		err := tx.QueryRow(
			`SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ?);`,
			m.table).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}

		if _, err := tx.Exec(m.copy); err != nil {
			return err
		}
		if _, err := tx.Exec(m.backup); err != nil {
			return err
		}
		logger.Info.Printf("migrated %s, old table kept as %s_backup\n", m.table, m.table)
	}

	return tx.Commit()
}

func seed(db *sql.DB) error {
	stmt, err := db.Prepare(`INSERT OR IGNORE INTO lockers (locker_id) VALUES (:id);`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 1; i <= 200; i++ {
		if _, err := stmt.Exec(sql.Named("id", fmt.Sprintf("ELW %03d", i))); err != nil {
			return err
		}
	}
	logger.Info.Println("seeded 200 lockers.")
	return nil
}
