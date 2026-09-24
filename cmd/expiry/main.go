package main

import (
	"fmt"

	"github.com/Uvic-ECSS/Lockers/internal/database"
	"github.com/Uvic-ECSS/Lockers/internal/email"
	"github.com/Uvic-ECSS/Lockers/internal/env"
	"github.com/Uvic-ECSS/Lockers/internal/logger"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		logger.Error.Fatal(err)
	}
	dbURL := fmt.Sprintf(
		"%s?authToken=%s",
		env.MustEnv("DATABASE_URL"),
		env.MustEnv("DATABASE_AUTH_TOKEN"))
	database.Connect(dbURL)
	email.Initialize()
}

// func queryExpiring() {

// }

func main() {
	//db, lock := database.Lock()
}
