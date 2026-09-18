package internal

import (
	"strconv"

	"github.com/Uvic-ECSS/ECSS-Lockers/internal/env"
)

var (
	Domain string
	Debug  bool
)

func Initialize() {
	Domain = env.Env("DOMAIN")
	Debug, _ = strconv.ParseBool(env.Env("DEBUG"))
}
