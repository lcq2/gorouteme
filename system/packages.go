package system

import (
	"bytes"
	"fmt"
	"gorouteme/templates"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

type packageInfo struct {
	Name    string
	Version string
	Size    string
	Summary string
}

var installedPackages []*packageInfo

func getInstalledPackages() {
	installedPackages = make([]*packageInfo, 0, 100)

	//out, err := exec.Command("dpkg-query", "-W", "-f", "'${binary:Package} ${Version} ${Installed-Size}\\n'").Output()
	out, err := exec.Command("rpm", "-qa", "--queryformat", "%{name};%{version}:%{release};%{size};%{summary}\\n").Output()
	if err != nil {
		log.Fatal(err)
	}

	b := bytes.NewBuffer(out)
	for {
		line, err := b.ReadString('\n')
		if err != nil {
			break
		}
		fields := strings.Split(strings.TrimSpace(line), ";")
		packageName := fields[0]
		packageVersion := fields[1]
		rawsize, err := strconv.Atoi(fields[2])
		if err != nil {
			rawsize = 0
		}
		packageSize := fmt.Sprintf("%.2f", float64(rawsize)/(1024.0*1024.0))
		packageSummary := fields[3]
		installedPackages = append(installedPackages, &packageInfo{Name: packageName, Version: packageVersion, Size: packageSize, Summary: packageSummary})
	}
}

func packagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		packagesData := map[string]interface{}{
			"NavBar":   true,
			"Packages": installedPackages,
		}
		err := templates.Manager().RenderView(w, "system_packages", packagesData)
		if err != nil {
			log.Fatal(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func init() {
	go getInstalledPackages()
}
