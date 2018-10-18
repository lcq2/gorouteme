package system

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gorouteme/templates"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type cpuStats struct {
	User    uint64
	Nice    uint64
	System  uint64
	Idle    uint64
	Iowait  uint64
	Irq     uint64
	Softirq uint64
}

type systemInfo struct {
	Hostname      string
	KernelName    string
	KernelRelease string
	KernelVersion string
	MachineName   string
	TotalMemory   string
	SwapTotal     string
	CPUCount      int
	CPUVendor     string
	CPUModel      string
}

type systemStatus struct {
	Uptime          string   `json:"uptime"`
	FreeMemory      string   `json:"freeMemory"`
	AvailableMemory string   `json:"availableMemory"`
	CachedMemory    string   `json:"cachedMemory"`
	SwapFree        string   `json:"swapFree"`
	CPUUsage        []string `json:"cpuUsage"`
}

const (
	daySecs  = 86400
	dayMins  = 1440
	dayHours = 24
	hourSecs = 3600
	hourMins = 60
	minSecs  = 60
)

var sysInfo systemInfo
var sysStatus systemStatus
var prevCPUStats []cpuStats

func parseProcMeminfo() map[string]int {
	meminfo, err := os.Open("/proc/meminfo")
	if err != nil {
		log.Fatal(err)
	}
	defer meminfo.Close()
	rd := bufio.NewReader(meminfo)
	mem := make(map[string]int)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			break
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		name := fields[0]
		value, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		mem[name[:len(name)-1]] = value
	}
	return mem
}

func parseProcStat() []cpuStats {
	totalCPUCount := sysInfo.CPUCount + 1
	procstat, err := os.Open("/proc/stat")
	if err != nil {
		log.Fatal(err)
	}
	defer procstat.Close()
	rd := bufio.NewReader(procstat)
	cpustats := make([]cpuStats, totalCPUCount)

	for i := 0; i < totalCPUCount; i++ {
		line, err := rd.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fields := strings.Fields(strings.TrimSpace(line))
		cpustats[i].User, _ = strconv.ParseUint(fields[1], 10, 64)
		cpustats[i].Nice, _ = strconv.ParseUint(fields[2], 10, 64)
		cpustats[i].System, _ = strconv.ParseUint(fields[3], 10, 64)
		cpustats[i].Idle, _ = strconv.ParseUint(fields[4], 10, 64)
		cpustats[i].Iowait, _ = strconv.ParseUint(fields[5], 10, 64)
		cpustats[i].Irq, _ = strconv.ParseUint(fields[6], 10, 64)
		cpustats[i].Softirq, _ = strconv.ParseUint(fields[7], 10, 64)
	}
	return cpustats
}

func toMibString(memsize int) string {
	return fmt.Sprintf("%.2f MiB", float64(memsize)/1024.0)
}

func updateSystemStatus() {
	data, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		log.Fatal(err)
	}
	uptime, err := strconv.ParseFloat(strings.Split(strings.TrimSpace(string(data)), " ")[0], 32)
	if err != nil {
		log.Fatal(err)
	}
	days := math.Floor(uptime / daySecs)
	hours := math.Floor(uptime/hourSecs) - days*dayHours
	mins := math.Floor(uptime/minSecs) - days*dayMins - hours*hourMins
	secs := math.Floor(uptime) - days*daySecs - hours*hourSecs - mins*minSecs
	sysStatus.Uptime = fmt.Sprintf("%d day(s), %d hour(s), %d minute(s), %d second(s)", int(days), int(hours), int(mins), int(secs))

	meminfo := parseProcMeminfo()
	sysStatus.FreeMemory = toMibString(meminfo["MemFree"])
	sysStatus.AvailableMemory = toMibString(meminfo["MemAvailable"])
	sysStatus.SwapFree = toMibString(meminfo["SwapFree"])
	sysStatus.CachedMemory = toMibString(meminfo["Cached"])

	curStats := parseProcStat()
	if len(sysStatus.CPUUsage) < sysInfo.CPUCount {
		sysStatus.CPUUsage = make([]string, sysInfo.CPUCount+1)
	}

	for i := 0; i < len(curStats); i++ {
		prevIdle := prevCPUStats[i].Idle + prevCPUStats[i].Iowait
		prevNonIdle := prevCPUStats[i].User + prevCPUStats[i].Nice + prevCPUStats[i].System + prevCPUStats[i].Irq + prevCPUStats[i].Softirq
		prevTotal := prevIdle + prevNonIdle

		curIdle := curStats[i].Idle + curStats[i].Iowait
		curNonIdle := curStats[i].User + curStats[i].Nice + curStats[i].System + curStats[i].Irq + curStats[i].Softirq
		curTotal := curIdle + curNonIdle

		total := curTotal - prevTotal
		idle := curIdle - prevIdle

		usage := ((float64(total) - float64(idle)) / float64(total)) * 100.0
		sysStatus.CPUUsage[i] = fmt.Sprintf("%.2f", usage)
	}

	prevCPUStats = curStats
	time.AfterFunc(time.Duration(5*time.Second), updateSystemStatus)
}
func statusHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		action := r.URL.Path[len("/system_status/"):]
		if action == "update" {
			json, err := json.Marshal(sysStatus)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(json)
		} else {
			statusData := map[string]interface{}{
				"NavBar":       true,
				"SystemInfo":   sysInfo,
				"SystemStatus": sysStatus,
			}
			err := templates.Manager().RenderView(w, "system_status", statusData)
			if err != nil {
				log.Fatal(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getSystemInfo() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "N/A"
	}
	sysInfo.Hostname = hostname

	out, err := exec.Command("uname", "-sri").Output()
	if err != nil {
		log.Fatal(err)
	}
	info := strings.Split(strings.TrimSpace(string(out)), " ")
	sysInfo.KernelName, sysInfo.KernelRelease, sysInfo.MachineName = info[0], info[1], info[2]

	out, err = exec.Command("uname", "-v").Output()
	if err != nil {
		log.Fatal(err)
	}
	sysInfo.KernelVersion = strings.TrimSpace(string(out))

	cpuinfo, err := os.Open("/proc/cpuinfo")
	if err != nil {
		log.Fatal(err)
	}
	defer cpuinfo.Close()
	rd := bufio.NewReader(cpuinfo)
	sysInfo.CPUCount = 0
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			break
		}
		if line[0] == '\n' {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		field := strings.TrimSpace(fields[0])
		if field == "processor" {
			sysInfo.CPUCount++
		} else if sysInfo.CPUVendor == "" || sysInfo.CPUModel == "" {
			value := strings.TrimSpace(fields[1])
			switch field {
			case "vendor_id":
				sysInfo.CPUVendor = value
			case "model name":
				sysInfo.CPUModel = value
			}
		}
	}

	mem := parseProcMeminfo()
	sysInfo.TotalMemory = toMibString(mem["MemTotal"])
	sysInfo.SwapTotal = toMibString(mem["SwapTotal"])
	prevCPUStats = parseProcStat()
}

func init() {
	getSystemInfo()
	time.AfterFunc(time.Duration(5*time.Second), updateSystemStatus)
}
