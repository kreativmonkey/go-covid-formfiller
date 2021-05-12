package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/browser"
)

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
		Title:      "Datenerfassung - Coronatest",
		Ldnr:       ldnr,
		Testcenter: cfg.Testcenter.City,
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
		signature_text = "Unterschrift der\nErziehungsberechtigten Person"
	}

	form := fillpdf.Form{
		"name":              fmt.Sprintf("%s %s", r.FormValue("fname"), r.FormValue("lname")),
		"signature":         fmt.Sprintf("%s %s", r.FormValue("fname"), r.FormValue("lname")),
		"signature_text":    signature_text,
		"bday":              bday.Format("02.01.2006"),
		"street_no":         r.FormValue("street"),
		"plz_city":          fmt.Sprintf("%s %s", r.FormValue("zip"), r.FormValue("city")),
		"date":              fmt.Sprintf("%s, %s", cfg.Testcenter.City, t.Format("02.01.2006")),
		"datetime_start":    fmt.Sprintf("%s, %s", cfg.Testcenter.City, t.Add(time.Minute*time.Duration(2)).Format("02.01.2006 15:04")),
		"datetime_end":      fmt.Sprintf("%s, %s", cfg.Testcenter.City, t.Add(time.Minute*time.Duration(17)).Format("02.01.2006 15:04")),
		"ldnr":              ldnr,
		"tc_plz_city":       cfg.Testcenter.Plz + " " + cfg.Testcenter.City,
		"tc_street_no":      cfg.Testcenter.Street,
		"tc_phone":          cfg.Testcenter.Phone,
		"tc_email":          cfg.Testcenter.Email,
		"test_manufacturer": cfg.Test.Hersteller,
		"test_pzn":          cfg.Test.Pzn,
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
