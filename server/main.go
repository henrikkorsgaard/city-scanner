package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/henrikkorsgaard/city-scanner/server/experiment"
)

var (
	gzipRE = regexp.MustCompile(`gzip`)
	gzRE   = regexp.MustCompile(`\.gz$`)
	files  = []string{filepath.Join("./client/templates", "main.tmpl")}
)

func main() {

	var dir string
	flag.StringVar(&dir, "dir", "./client/static/", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()

	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))

	r.HandleFunc("/", overviewHandler).Methods("GET")
	r.HandleFunc("/create", createHandler).Methods("GET", "POST")
	r.HandleFunc("/validate/{key}/{value}", validateHandler).Methods("GET")
	r.HandleFunc("/experiment/{experiment}", experimentHandler).Methods("GET")
	r.HandleFunc("/api/", apiHandler).Methods("GET", "POST")

	log.Fatal(http.ListenAndServe(":2488", r))

}

func overviewHandler(w http.ResponseWriter, r *http.Request) {

	files = append(files, filepath.Join("./client/templates", "overview.tmpl"))
	tmpl, err := template.ParseFiles(files...)

	if err != nil {
		// Log the detailed error
		log.Println(err.Error())
		// Return a generic "Internal Server Error" message
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "overview", nil); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		jsonData, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal("Error reading the body", err)
		}

		experiment, exists, err := experiment.NewExperiment(jsonData)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(exists)
		//exists should be handled in valudation
		fmt.Println(experiment)

	} else {
		files = append(files, filepath.Join("./client/templates", "create.tmpl"))

		tmpl, err := template.ParseFiles(files...)

		if err != nil {
			// Log the detailed error
			log.Println(err.Error())
			// Return a generic "Internal Server Error" message
			http.Error(w, http.StatusText(500), 500)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "create", nil); err != nil {
			log.Println(err.Error())
			http.Error(w, http.StatusText(500), 500)
		}
	}
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	switch key := vars["key"]; key {
	case "name":
		value := vars["value"]
		for _, expName := range experiment.AllExperimentNames {
			if expName == value {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Exists\n"))
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404\n"))

		return
	default:
		fmt.Println("Recieved unknown validation key!")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404\n"))

		return
	}
}

func experimentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["experiment"]
	e, err := experiment.GetExperiment(slug)
	fmt.Println(e)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404\n"))
	} else {
		w.Write([]byte("good\n"))
	}

}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Gorilla!\n"))
}

func dataHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	}
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
