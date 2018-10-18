package system

import (
	"fmt"
	"gorouteme/templates"
	"log"
	"net/http"
	"os/exec"
)

func shutdown() {
	fmt.Println("system.shutdown")
	cmd := exec.Command("/sbin/shutdown", "-h", "now")
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func reboot() {
	fmt.Println("system.reboot")
	cmd := exec.Command("/sbin/shutdown", "-r", "now")
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		shutdownData := map[string]interface{}{
			"Timeout": 15,
			"NavBar":  true,
		}
		err := templates.Manager().RenderView(w, "system_shutdown", shutdownData)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		action := r.URL.Path[len("/system_shutdown/"):]
		switch action {
		case "shutdown":
			go shutdown()
		case "reboot":
			go reboot()
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
