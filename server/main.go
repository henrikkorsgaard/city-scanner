package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/gorilla/mux"
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

	r.HandleFunc("/", overview).Methods("GET")
	r.HandleFunc("/create", create).Methods("GET", "POST")
	r.HandleFunc("/experiment/{experiment}", experiment).Methods("GET")
	r.HandleFunc("/api", api).Methods("GET", "POST")

	log.Fatal(http.ListenAndServe(":2488", r))
}

func overview(w http.ResponseWriter, r *http.Request) {

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

func create(w http.ResponseWriter, r *http.Request) {

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

func experiment(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Gorilla!\n"))
}

func api(w http.ResponseWriter, r *http.Request) {
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
