package main

import (
	"crypto/subtle"
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

	r.HandleFunc("/", BasicAuth(overviewHandler)).Methods("GET")
	r.HandleFunc("/create", BasicAuth(createHandler)).Methods("GET", "POST")
	r.HandleFunc("/validate/{key}/{value}", validateHandler).Methods("GET")
	r.HandleFunc("/experiment/{experiment}", BasicAuth(experimentHandler)).Methods("GET")
	r.HandleFunc("/experiment/{experiment}/node/{id}", nodeHandler).Methods("GET", "POST")
	r.HandleFunc("/experiment/{experiment}/configurationfile.config", experimentConfigurationFileHandler).Methods("GET")
	r.HandleFunc("/api/", apiHandler).Methods("GET", "POST")
	//r.HandleFunc("/favicon.ico", faviconHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":2488", r))

}

func nodeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Println("akhdsa")
	//slug := vars["experiment"]
	//e, err := experiment.GetExperiment(slug)

	//if post -> save node and return id
	//if get return node
	if r.Method == "POST" {
		nid := vars["id"]
		//jsonData, err := ioutil.ReadAll(r.Body)
		/*
			if err != nil {
				log.Fatal("Error reading the body", err)
			}*/
		fmt.Println(nid)
		/*
			n, err := e.GetNode(nid)

			fmt.Println(n, err)
		*/
		/*
			jsonData, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Fatal("Error reading the body", err)
			}

			node, _, err := node.NewExperiment(jsonData)
			if err != nil {
				log.Fatal(err)
			}*/
		/*
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("/experiment/" + experiment.Slug))
		*/

	} else {
		//return node json
	}
}

func overviewHandler(w http.ResponseWriter, r *http.Request) {

	files = append(files, filepath.Join("./client/templates", "overview.tmpl"))
	tmpl, err := template.ParseFiles(files...)
	experiments := experiment.GetAllExperiments()
	if err != nil {
		// Log the detailed error
		log.Println(err.Error())
		// Return a generic "Internal Server Error" message
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "overview", experiments); err != nil {
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

		experiment, _, err := experiment.NewExperiment(jsonData)
		if err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("/experiment/" + experiment.Slug))

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

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404\n"))
	} else {
		files = append(files, filepath.Join("./client/templates", "experiment.tmpl"))
		tmpl, err := template.ParseFiles(files...)

		if err != nil {
			// Log the detailed error
			log.Println(err.Error())
			// Return a generic "Internal Server Error" message
			http.Error(w, http.StatusText(500), 500)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "experiment", e); err != nil {
			log.Println(err.Error())
			http.Error(w, http.StatusText(500), 500)
		}
	}

}

func experimentConfigurationFileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["experiment"]

	e, err := experiment.GetExperiment(slug)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404\n"))
	} else {
		w.Header().Set("Content-Disposition", "attachment; filename=city-scanner.config")
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

		w.Write([]byte(e.GenerateConfigurationFile()))
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

func BasicAuth(handler http.HandlerFunc) http.HandlerFunc {
	username := "aarhus"
	password := "city"
	return func(w http.ResponseWriter, r *http.Request) {

		user, pass, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Please enter password"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}

		handler(w, r)
	}
}

/*
func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join(publicPath, "./images/favicon.ico"))
}*/
