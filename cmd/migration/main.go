package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/Uvic-ECSS/ECSS-Lockers/internal/database"
	"github.com/Uvic-ECSS/ECSS-Lockers/internal/env"
	"github.com/Uvic-ECSS/ECSS-Lockers/internal/logger"
)

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

	schema, err := os.ReadFile("internal/database/schema.sql")
	if err != nil {
		logger.Error.Fatal(err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		logger.Error.Fatal(err)
	}

	logger.Info.Println("created schema.")

	var oldLockerTableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM sqlite_master
			WHERE type = 'table' AND name = 'locker'
		);`,
	).Scan(&oldLockerTableExists)
	if err != nil {
		logger.Error.Fatal(err)
	}

	if oldLockerTableExists {
		if _, err := db.Exec(`
			INSERT OR IGNORE INTO lockers (locker_id)
			SELECT id
			FROM locker
			WHERE id IS NOT NULL AND id <> '';
		`); err != nil {
			logger.Error.Fatal(err)
		}

		logger.Info.Println("migrated lockers from locker(id) to lockers(locker_id)")
	}

	var oldRegistrationTableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM sqlite_master
			WHERE type = 'table' AND name = 'registration'
		);`,
	).Scan(&oldRegistrationTableExists)
	if err != nil {
		logger.Error.Fatal(err)
	}

	if oldRegistrationTableExists {
		if _, err := db.Exec(`
			INSERT OR IGNORE INTO locker_registrations (
				locker_id,
				user_email,
				user_name,
				expiry_date,
				expiry_email_sent
			)
			SELECT locker, user, name, expiry, expiryEmailSent
			FROM registration;
		`); err != nil {
			logger.Error.Fatal(err)
		}

		logger.Info.Println("migrated registrations from registration to locker_registrations")
	}

	logger.Info.Println("seeding 200 lockers..")
	// eeehhh i'm not proud of how this is being done but
	// database/sql does not support array type for query
	// arg out of the box :(
	for i := 0; i < 200; i++ {
		locker := fmt.Sprintf("ELW %03d", i+1)

		stmt, err := db.Prepare(`INSERT INTO lockers (locker_id) VALUES (:id);`)
		if err != nil {
			logger.Error.Fatal(err)
		}

		_, err = stmt.Exec(sql.Named("id", locker))
		if err != nil {
			logger.Error.Printf("error seeding locker %s:\n%v", locker, err)
		}
	}

	logger.Info.Println("Done")
}
