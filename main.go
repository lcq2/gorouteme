package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gorouteme/system"
	"gorouteme/utils"
	"html/template"
	"io"
	"io/ioutil"
	"os/exec"
	"strconv"
	"time"
	//	"io/ioutil"
	"golang.org/x/crypto/pbkdf2"
	"gorouteme/session"
	"gorouteme/templates"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Config struct {
	Login struct {
		Salt     string `json:"salt"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	Session struct {
		CookieName  string `json:"cookieName"`
		MaxLifeTime int64  `json:"maxLifeTime"`
	}
	System struct {
		ShutdownTimeout string `json:"shutdownTimeout"`
	}
	BindAddress string `json:"bindAddress"`
	BindPort    int    `json:"bindPort"`
	WebFolder   string `json:"webFolder"`
}

var mainConfig Config

func staticServeFile(w http.ResponseWriter, r *http.Request) {
	dir, file := filepath.Split(r.URL.Path)
	fullpath := filepath.Join(mainConfig.WebFolder, "static", dir, file)
	st, err := os.Stat(fullpath)
	if err != nil {
		log.Fatal(err)
	}
	staticFile, err := os.Open(fullpath)
	defer staticFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Cache-Control", "no-cache, no-store")
	http.ServeContent(w, r, file, st.ModTime(), staticFile)
}

func login(w http.ResponseWriter, r *http.Request) {
	s := session.Manager().SessionGet(r)
	switch r.Method {
	case http.MethodGet:
		if s != nil && s.UserLoggedIn() {
			http.Redirect(w, r, "/status/", http.StatusSeeOther)
			return
		}

		b := make([]byte, 32)
		rand.Read(b)
		h := sha256.New()
		h.Write(b)
		io.WriteString(h, strconv.FormatInt(time.Now().Unix(), 10))
		token := hex.EncodeToString(h.Sum(nil))
		s := session.Manager().SessionNew(w, r)
		s.Set("loginToken", token)

		loginData := struct {
			NavBar     bool
			LoginToken string
		}{
			false,
			token,
		}
		err := templates.Manager().RenderView(w, "login", loginData)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		if s == nil {
			http.Error(w, "Invalid session", http.StatusForbidden)
			return
		}
		r.ParseForm()
		if s.Get("loginToken") != r.Form.Get("loginToken") {
			http.Error(w, "Invalid token", http.StatusForbidden)
			return
		}
		inputUsername := template.HTMLEscapeString(r.Form.Get("inputUsername"))
		inputPassword := template.HTMLEscapeString(r.Form.Get("inputPassword"))

		dk := hex.EncodeToString(pbkdf2.Key([]byte(inputPassword), []byte(mainConfig.Login.Salt), 100000, 32, sha256.New))
		if inputUsername == mainConfig.Login.Username && dk == mainConfig.Login.Password {
			s.Set("loggedIn", true)
			s.Set("username", inputUsername)
			http.Redirect(w, r, "/status/", http.StatusSeeOther)
		} else {
			s.Flush()
			session.Manager().SessionDestroy(w, r)
			http.Error(w, "Invalid username or password", http.StatusForbidden)
		}

	case http.MethodHead:
		w.Header().Set("Server", "Alive")
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		session.Manager().SessionDestroy(w, r)
		http.Redirect(w, r, "/login/", http.StatusSeeOther)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func LoadConfig(file string) (config Config) {
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return
}

func unpackWebArchive(filename string) (string, error) {
	extension := filepath.Ext(filename)
	tmpdir, err := ioutil.TempDir("", "grm")
	if err != nil {
		return "", err
	}
	fmt.Printf("Unpacking web archive %s to %s\n", filename, tmpdir)

	var cmd *exec.Cmd = nil
	switch extension {
	case ".gz":
		fallthrough
	case ".tgz":
		cmd = exec.Command("tar", "-xzf", filename, "-C", tmpdir)
	case ".bz2":
		cmd = exec.Command("tar", "-xjf", filename, "-C", tmpdir)
	case ".xz":
		cmd = exec.Command("tar", "-xJf", filename, "-C", tmpdir)
	case ".zip":
		cmd = exec.Command("unzip", filename, "-d", tmpdir)
	}
	fmt.Printf("Will run: %s %v\n", cmd.Path, cmd.Args)
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	return tmpdir, nil
}

func main() {
	mainConfig = LoadConfig("cfg.json")

	if mainConfig.BindAddress == "0.0.0.0" {
		log.Fatal("Sorry, I will not bind to 0.0.0.0. Use port redirection for external access.")
	}
	webFolder := mainConfig.WebFolder

	st, err := os.Stat(webFolder)
	if err != nil {
		log.Fatal(err)
	}
	if !st.IsDir() {
		webFolder, err = unpackWebArchive(webFolder)
		if err != nil {
			log.Fatal(err)
		}
	}

	session.InitSessionManager(mainConfig.Session.CookieName, mainConfig.Session.MaxLifeTime)
	templates.InitTemplateManager(filepath.Join(webFolder, "templates"))

	http.HandleFunc("/css/", staticServeFile)
	http.HandleFunc("/scripts/", staticServeFile)

	http.HandleFunc("/login/", login)
	http.HandleFunc("/logout/", utils.LoginRequired(logout))

	system.RegisterHandlers()

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", mainConfig.BindAddress, mainConfig.BindPort), nil))
}
