package main

import "fmt"

var (
	appNameLong  = "Prometheus Import Manager"
	appNameShort = "pim"
	appVersion   = "v0.1.0"
)

func Version() string {
	return fmt.Sprintf("%s - %s %s", appNameShort, appNameLong, appVersion)
}
