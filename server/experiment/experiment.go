package experiment

import (
	mrand "math/rand"
	"runtime"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"

	"log"
	"os"
	"path/filepath"

	"time"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

/*
WE NEED TO MAKE THE PUBLIC KEY PART OF THE EXPERIMENT INITIALISATION
*/

type Experiment struct {
	ID               int     //
	Slug             string  //Slug used to show it on the platform + date
	Name             string  `json:"name"`  //Name end-user identifier
	Email            string  `json:"email"` //Email to the primary researcher
	Latitude         float64 `json:"lat"`   //Latitude coords to the map center of the experiment
	Longitude        float64 `json:"lng"`
	PrivateKey       []byte  //marshalled private key -- see https://stackoverflow.com/questions/13555085/save-and-load-crypto-rsa-privatekey-to-and-from-the-disk
	PublicKeyPEM     []byte  //Public key marshalled
	DatabaseFileName string
	NodeSalt         string //We need salt at the nodes in order to obfuscate the mac addresses (it is a finite spaces)
	Nodes            []Node
	Active           bool
}

type Node struct {
	ID          int
	Description string
	Location    string
	Latitude    float64
	Longitude   float64
	Readings    []Reading
}

type Reading struct {
	ID        int
	NodeID    int
	DeviceID  string
	Signal    int
	Timestamp time.Time
}

var (
	experimentDB       *sql.DB
	DBDir              string
	DBExperimentsDir   string
	AllExperimentNames = []string{}
	linebreak          = "\n"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
)

func init() {
	var err error
	DBDir, err = filepath.Abs("client/databases")
	DBExperimentsDir, err = filepath.Abs("client/databases/experiments")
	//If the database file does not exist sql.Open(...) will fail, so we create the db first
	if _, err = os.Stat(filepath.Join(DBDir, "experiments.sqlite.db")); os.IsNotExist(err) {
		_, err = os.OpenFile(filepath.Join(DBDir, "experiments.sqlite.db"), os.O_RDONLY|os.O_CREATE, 0666)
	}
	//Open database (not really, but preparing for lazy open on first request)
	experimentDB, err = sql.Open("sqlite3", filepath.Join(DBDir, "experiments.sqlite.db"))
	if err != nil {
		log.Fatal(err)
	}

	//Setting up the table that contains the experiments
	statement, err := experimentDB.Prepare("CREATE TABLE IF NOT EXISTS experiments (id INTEGER PRIMARY KEY, slug TEXT, name TEXT, email TEXT, latitude REAL, longitude REAL, privatekey BLOB, publickeyPEM BLOB, nodesalt TEXT, databasefilename TEXT, active INT)")
	if err != nil {
		log.Fatal(err)
	}
	statement.Exec()

	AllExperimentNames = getAllExperimentNames()
	fmt.Println(AllExperimentNames)

	if runtime.GOOS == "windows" {
		linebreak = "\r\n"
	}
}

func NewExperiment(jsonData []byte) (e Experiment, exists bool, err error) {

	e = Experiment{}
	err = json.Unmarshal(jsonData, &e)
	if err != nil {
		return
	}

	//If the database already exists it should return here
	if exist := experimentExists(e.Name); exist {
		exists = true
		return
	}

	e.Slug = strings.Replace(e.Name, " ", "-", -1)
	e.Active = true
	//https://medium.com/@raul_11817/golang-cryptography-rsa-asymmetric-algorithm-e91363a2f7b3
	privkey, err := rsa.GenerateKey(rand.Reader, 1024)
	//https://stackoverflow.com/questions/13555085/save-and-load-crypto-rsa-privatekey-to-and-from-the-disk
	e.PrivateKey = x509.MarshalPKCS1PrivateKey(privkey)
	var pubkeyMarshalled []byte
	pubkeyMarshalled, err = x509.MarshalPKIXPublicKey(&privkey.PublicKey)
	if err != nil {
		return
	}

	e.PublicKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubkeyMarshalled,
	})

	e.DatabaseFileName = e.Slug + ".sqlite.db"
	if _, err = os.Stat(filepath.Join(DBExperimentsDir, e.DatabaseFileName)); os.IsNotExist(err) {
		_, err = os.OpenFile(filepath.Join(DBExperimentsDir, e.DatabaseFileName), os.O_RDONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println("1")
			return
		}
	}

	e.NodeSalt = RandStringBytes(32) //Short salt to obfuscate mac addresses

	var statement *sql.Stmt
	statement, err = experimentDB.Prepare("INSERT INTO experiments (slug, name, email, latitude, longitude, privatekey, publickeyPEM, nodesalt, databasefilename) VALUES (?,?,?,?,?,?,?,?,?)")

	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = statement.Exec(e.Slug, e.Name, e.Email, e.Latitude, e.Longitude, e.PrivateKey, e.PublicKeyPEM, e.NodeSalt, e.DatabaseFileName)

	if err != nil {
		fmt.Println("ældjas")
		fmt.Println(err)
		return
	}

	var db *sql.DB
	db, err = sql.Open("sqlite3", filepath.Join(DBExperimentsDir, e.DatabaseFileName))

	if err != nil {
		fmt.Println("2")
		return
	}
	//We don't mind this being closed before returning experiment as we will work on this per experiment case
	defer db.Close()

	//Setting up table for nodes

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS nodes (id INTEGER PRIMARY KEY, description TEXT, location TEXT, latitude REAL, longitude REAL)")

	if err != nil {
		fmt.Println("3")
		return
	}
	statement.Exec()

	//Setting up table for readings
	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS readings (id INTEGER PRIMARY KEY, nodeID INT, deviceID TEXT, signal INT, Timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)")

	if err != nil {
		fmt.Println("4")
		return
	}
	statement.Exec()
	AllExperimentNames = getAllExperimentNames()

	return
}

//Public so careful!
func GetExperiment(slug string) (e Experiment, err error) {
	e = Experiment{}
	statement := `SELECT name, slug, email, latitude, longitude, nodesalt, publickeypem FROM experiments WHERE slug=$1;`

	row := experimentDB.QueryRow(statement, slug)
	err = row.Scan(&e.Name, &e.Slug, &e.Email, &e.Latitude, &e.Longitude, &e.NodeSalt, &e.PublicKeyPEM)
	return
}

func GetAllExperiments() (experiments []Experiment) {
	statement := `SELECT name,slug FROM experiments;`
	var name string
	var slug string

	rows, err := experimentDB.Query(statement)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&name, &slug)
		if err != nil {
			log.Fatal(err)
		}

		experiments = append(experiments, Experiment{Name: name, Slug: slug})
	}

	return experiments
}

//Public so careful!
func (e *Experiment) GenerateConfigurationFile() (configs []byte) {
	var iniString strings.Builder
	iniString.WriteString("#This configuration file is generated by city-scanner." + linebreak)
	iniString.WriteString("#Please consult the online guide on how to configure the sensing node." + linebreak + linebreak)
	iniString.WriteString("#The network configuration section supports add network information for the sensing node to connect to the server for data upload." + linebreak)
	iniString.WriteString("[Experiment Information]" + linebreak)
	iniString.WriteString("name=")
	iniString.WriteString(e.Name)
	iniString.WriteString(linebreak)
	iniString.WriteString("webpage=")
	iniString.WriteString("") //TODO: generate based on server information
	iniString.WriteString(linebreak)
	iniString.WriteString("email_contact=")
	iniString.WriteString(e.Email)
	iniString.WriteString(linebreak + linebreak)

	iniString.WriteString("[Network Configuration]" + linebreak)
	iniString.WriteString("ssid=" + linebreak)
	iniString.WriteString("password=" + linebreak)
	iniString.WriteString("[Server information]" + linebreak)
	iniString.WriteString("server_url=" + linebreak) //TODO add this server
	iniString.WriteString("upload_frequency_seconds=600" + linebreak + linebreak)
	iniString.WriteString("[Security]" + linebreak)
	iniString.WriteString("salt=")
	fmt.Println(e.NodeSalt)
	iniString.WriteString(e.NodeSalt)
	iniString.WriteString(linebreak)
	iniString.WriteString(string(e.PublicKeyPEM))
	configs = []byte(iniString.String())
	return
}

func getAllExperimentNames() []string {
	names := []string{}
	statement := `SELECT name FROM experiments;`
	var name string
	rows, err := experimentDB.Query(statement)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&name)
		if err != nil {
			log.Fatal(err)
		}

		names = append(names, name)
	}

	return names
}

func experimentExists(name string) bool {
	//we need to check for this elsewhere
	statement := `SELECT id FROM experiments WHERE name=$1;`
	var id int

	row := experimentDB.QueryRow(statement, name)
	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		return false
	case nil:
		return true
	default:
		panic(err)
	}
}

//https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[mrand.Intn(len(letterBytes))]
	}
	return string(b)
}
