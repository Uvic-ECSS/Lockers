package internal

import (
	"strconv"

	"github.com/parsa222/ECSS-Lockers/internal/env"
)

var (
	Domain string
	Debug  bool
)

func Initialize() {
	Domain = env.Env("DOMAIN")
	Debug, _ = strconv.ParseBool(env.Env("DEBUG"))
}
