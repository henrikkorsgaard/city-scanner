package experiment

import (
	mrand "math/rand"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"strings"

	"log"
	"os"
	"path/filepath"

	"time"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Experiment struct {
	ID               int     //
	Slug             string  //Slug used to show it on the platform + date
	Name             string  `json:"name"`  //Name end-user identifier
	Email            string  `json:"email"` //Email to the primary researcher
	Latitude         float64 `json:"lat"`   //Latitude coords to the map center of the experiment
	Longitude        float64 `json:"lng"`
	PrivateKey       []byte  //marshalled private key -- see https://stackoverflow.com/questions/13555085/save-and-load-crypto-rsa-privatekey-to-and-from-the-disk
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
	experimentDB     *sql.DB
	DBDir            string
	DBExperimentsDir string
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

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
	statement, err := experimentDB.Prepare("CREATE TABLE IF NOT EXISTS experiments (id INTEGER PRIMARY KEY, slug TEXT, name TEXT, email TEXT, latitude REAL, longitude REAL, privatekey BLOB, nodesalt TEXT, databasefilename TEXT, active INT)")
	if err != nil {
		log.Fatal(err)
	}
	statement.Exec()
}

func NewExperiment(jsonData []byte) (e Experiment, err error) {

	e = Experiment{}
	err = json.Unmarshal(jsonData, &e)
	if err != nil {
		return
	}
	fmt.Println(e.Name)
	//If the database already exists it should return here
	if exist := experimentExists(e.Name); exist {
		fmt.Print("already exist!")
		err = fmt.Errorf("Record already exist")
		return
	}

	e.Slug = strings.Replace(e.Name, " ", "-", -1)
	//https://medium.com/@raul_11817/golang-cryptography-rsa-asymmetric-algorithm-e91363a2f7b3
	privkey, err := rsa.GenerateKey(rand.Reader, 1024)
	//https://stackoverflow.com/questions/13555085/save-and-load-crypto-rsa-privatekey-to-and-from-the-disk
	e.PrivateKey = x509.MarshalPKCS1PrivateKey(privkey)

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
	statement, err = experimentDB.Prepare("INSERT INTO experiments (slug, name, email, latitude, longitude, privatekey, nodesalt, databasefilename) VALUES (?,?,?,?,?,?,?,?)")

	if err != nil {
		fmt.Println(err)
		return
	}
	var result sql.Result
	result, err = statement.Exec(e.Slug, e.Name, e.Email, e.Latitude, e.Longitude, e.PrivateKey, e.NodeSalt, e.DatabaseFileName)

	if err != nil {
		fmt.Println("ældjas")
		fmt.Println(err)
		return
	}

	fmt.Println(result)

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

	//we also need to generate a config file and a private keys
	log.Printf("file: %v\n", e)
	return
}

//Public so careful!
func GetExperiment(name string) (e Experiment, err error) {
	e = Experiment{}
	statement := `SELECT name, slug, email, latitude, longitude FROM experiments WHERE name=$1;`

	row := experimentDB.QueryRow(statement, name)
	err = row.Scan(&e.Name, &e.Slug, &e.Email, &e.Latitude, &e.Longitude)
	return
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