package profile

import (
	"fmt"
	"os"
	"strings"
)

type ProfileType string

var Current = DEV // dev profile as default

const (
	DEV  ProfileType = "DEV"
	TEST ProfileType = "TEST"
	PROD ProfileType = "PROD"
)

func InitProfile() {
	switch strings.ToUpper(os.Getenv("PROFILE")) {
	case "DEV":
		Current = DEV
	case "TEST":
		Current = TEST
	case "PROD":
		Current = PROD
	}
	fmt.Printf("Current profile: %s\n", Current)
}
