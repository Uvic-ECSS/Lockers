package internal

import (
	"github.com/parsa222/ECSS-Lockers/internal/env"
)

var (
	Domain string
)

func Initialize() {
	Domain = env.Env("DOMAIN")
}
