package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/browser"

	"github.com/desertbit/fillpdf"
	"gopkg.in/yaml.v2"
)

var (
	tpl *template.Template
	cfg *Config
)

type Page struct {
	Title string
	Ldnr  string
}

type Config struct {
	Server struct {
		// Port is the local machine TCP Port to bind the HTTP Server to
		Port string `yaml:"port"`

		// Host is the local machine IP Address to bind the HTTP Server to
		Host string `yaml:"host"`

		SavePath string `yaml:"save_path"`

		Timeout struct {
			// Server is the general server timeout to use
			// for graceful shutdowns
			Server time.Duration `yaml:"server"`

			// Write is the amount of time to wait until an HTTP server
			// write opperation is cancelled
			Write time.Duration `yaml:"write"`

			// Read is the amount of time to wait until an HTTP server
			// read operation is cancelled
			Read time.Duration `yaml:"read"`

			// Read is the amount of time to wait
			// until an IDLE HTTP session is closed
			Idle time.Duration `yaml:"idle"`
		} `yaml:"timeout"`
	} `yaml:"server"`
	Testcenter struct {
		Street string `yaml:"street"`
		Plz    string `yaml:"plz"`
		City   string `yaml:"city"`
		Phone  string `yaml:"phone"`
		Email  string `yaml:"email"`
	} `yaml:"testcenter"`
	Ldnr struct {
		Prefix    string `yaml:"prefix"`
		Counter   int    `yaml:"counter"`
		NumLength int    `yaml:"numlength"`
	} `yaml:"ldnr"`
	Test struct {
		Hersteller string `yaml:"hersteller"`
		Pzn        string `yaml:"pzn"`
	} `yaml:"test"`
}

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}

// ParseFlags will create and parse the CLI flags
// and return the path to be used elsewhere
func ParseFlags() (string, error) {
	// String that contains the configured configuration path
	var configPath string

	// Set up a CLI flag called "-config" to allow users
	// to supply the configuration file
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")

	// Actually parse the flags
	flag.Parse()

	// Validate the path first
	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	// Return the configuration path
	return configPath, nil
}

// NewRouter generates the router used in the HTTP Server
func NewRouter() *http.ServeMux {
	// Create router and define routes and return that router
	router := http.NewServeMux()

	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/fillform", fillForm)
	return router
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Call Index")
	ldnr := cfg.Ldnr.Prefix + fmt.Sprintf("%0*d", cfg.Ldnr.NumLength, cfg.Ldnr.Counter)
	p := Page{
		Title: "Datenerfassung - Coronatest",
		Ldnr:  ldnr,
	}
	err := tpl.ExecuteTemplate(w, "index.html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func fillForm(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Call fillForm")
	if r.Method == "GET" {
		fmt.Println("Its a GET Request")
		return
	}

	r.ParseForm()

	t := time.Now()

	// Fill the Forms with the following Form map
	ldnr := r.FormValue("ldnr")

	// Converte the bday string into time
	bday, err := time.Parse("2006-01-02", r.FormValue("bday"))
	if err != nil {
		log.Fatal(err)
	}
	signature_text := "Unterschrift der zu testenden Person"
	if !validateAge(bday, 18) {
		signature_text = "Unterschrift der Erziehungsberechtigten Person\n"
	}

	form := fillpdf.Form{
		"name":            fmt.Sprintf("%s %s", r.FormValue("fname"), r.FormValue("lname")),
		"signature":       fmt.Sprintf("%s %s", r.FormValue("fname"), r.FormValue("lname")),
		"signature_text":  signature_text,
		"bday":            bday.Format("02.01.2006"),
		"street_no":       r.FormValue("street"),
		"plz_city":        fmt.Sprintf("%s %s", r.FormValue("zip"), r.FormValue("city")),
		"date":            fmt.Sprintf("%s, %s", cfg.Testcenter.City, t.Add(time.Minute*time.Duration(2)).Format("02.01.2006")),
		"datetime_start":  t.Format("02.01.2006 15:04"),
		"datetime_end":    fmt.Sprintf("%s, %s", cfg.Testcenter.City, t.Add(time.Minute*time.Duration(15)).Format("02.01.2006 15:04")),
		"ldnr":            ldnr,
		"tc_plz_city":     cfg.Testcenter.Plz + " " + cfg.Testcenter.City,
		"tc_street_no":    cfg.Testcenter.Street,
		"tc_phone":        cfg.Testcenter.Phone,
		"tc_email":        cfg.Testcenter.Email,
		"test_hersteller": cfg.Test.Hersteller,
		"test_pzn":        cfg.Test.Pzn,
	}

	fmt.Println("Fillpdf with ", form)
	// Filling form using pdftk
	// save the file as ldnr.Tag + ldnr.StartCounter + pdf Example: OO-030.pdf
	filepath := cfg.Server.SavePath + strings.Trim(ldnr, "#") + ".pdf"
	err = fillpdf.Fill(form, "formular.pdf", filepath, fillpdf.Options{true, true})
	if err != nil {
		log.Fatal(err)
	}

	// Open Dowload from file
	//w.Header().Set("Content-Disposition", "attachment; filename="+filepath)
	//w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	//http.ServeFile(w, r, filepath)
	browser.OpenURL(filepath)
	cfg.Ldnr.Counter += 1
	updateConfig(*cfg)
	http.Redirect(w, r, r.Header.Get("Referer"), 302)
}

// validate birthday
func validateAge(birthdate time.Time, checkage int) bool {
	today := time.Now()
	today = today.In(birthdate.Location())
	ty, tm, td := today.Date()
	today = time.Date(ty, tm, td, 0, 0, 0, 0, time.UTC)
	by, bm, bd := birthdate.Date()
	birthdate = time.Date(by, bm, bd, 0, 0, 0, 0, time.UTC)
	if today.Before(birthdate) {
		return false
	}
	age := ty - by
	anniversary := birthdate.AddDate(age, 0, 0)
	if anniversary.After(today) {
		age--
	}

	if age < checkage {
		fmt.Println("Ist noch NICHT 18 Jahre!!!")
		return false
	}
	return true
}

// Run will run the HTTP Server
func (config Config) Run() {
	// Set up a channel to listen to for interrupt signals
	var runChan = make(chan os.Signal, 1)

	// Set up a context to allow for graceful server shutdowns in the event
	// of an OS interrupt (defers the cancel just in case)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		config.Server.Timeout.Server,
	)
	defer cancel()

	// Define server options
	server := &http.Server{
		Addr:         config.Server.Host + ":" + config.Server.Port,
		Handler:      NewRouter(),
		ReadTimeout:  config.Server.Timeout.Read * time.Second,
		WriteTimeout: config.Server.Timeout.Write * time.Second,
		IdleTimeout:  config.Server.Timeout.Idle * time.Second,
	}
	tpl = template.Must(template.ParseGlob("./views/*.html"))

	// Handle ctrl+c/ctrl+x interrupt
	// That is only for Linux Code!!
	// signal.Notify(runChan, os.Interrupt, syscall.SIGTSTP)

	// Alert the user that the server is starting
	log.Printf("Server is starting on %s\n", server.Addr)

	log.Printf(cfg.Ldnr.Prefix)
	// Run the server on a new goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				// Normal interrupt operation, ignore
			} else {
				log.Fatalf("Server failed to start due to err: %v", err)
			}
		}
	}()

	browser.OpenURL("http://localhost:" + config.Server.Port)
	// Block on this channel listeninf for those previously defined syscalls assign
	// to variable so we can let the user know why the server is shutting down
	interrupt := <-runChan

	// If we get one of the pre-prescribed syscalls, gracefully terminate the server
	// while alerting the user
	log.Printf("Server is shutting down due to %+v\n", interrupt)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server was unable to gracefully shutdown due to err: %+v", err)
	}
}

// Func main should be as small as possible and do as little as possible by convention
func main() {
	// Generate our config based on the config supplied
	// by the user in the flags
	cfgPath, err := ParseFlags()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err = NewConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	// Run the server
	cfg.Run()
}

func updateConfig(config Config) {
	log.Printf("Laufende Nummer erhöht zu: " + fmt.Sprintf("%x", config.Ldnr.Counter))
	d, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = ioutil.WriteFile("config.yml", d, 0644)
	if err != nil {
		log.Fatal(err)
	}

}

/**
* To Download files:
* https://stackoverflow.com/questions/24116147/how-to-download-file-in-browser-from-go-server
**/
