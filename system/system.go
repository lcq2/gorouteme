package system

import (
	"gorouteme/utils"
	"net/http"
)

func RegisterHandlers() {
	http.HandleFunc("/system_status/", statusHandler)
	http.HandleFunc("/system_packages/", packagesHandler)
	http.HandleFunc("/system_shutdown/", utils.LoginRequired(shutdownHandler))
}
