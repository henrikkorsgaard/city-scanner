package experiment

import (
	"crypto/rand"
	"crypto/rsa"
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
	ID         int     //Db id
	Slug       string  //Slug used to show it on the platform
	Name       string  `json:"name"`  //Name end-user identifier
	Email      string  `json:"email"` //Email to the primary researcher
	Latitude   float64 `json:"lat"`   //Latitude coords to the map center of the experiment
	Longitude  float64 `json:"lng"`
	PrivateKey rsa.PrivateKey
	Database   os.File
	NodeSalt   string //We need salt at the nodes in order to obfuscate the mac addresses (it is a finite spaces)
	Nodes      []Node
}

type Node struct {
	ID          int
	Description string
	Location    string
	Latitude    float64
	Longitude   float64
	PublicKey   rsa.PublicKey
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
		fmt.Println("0") //FATAL
		panic(err)
	}
	//defer experimentDB.Close()

	//Setting up the table that contains the experiments
	statement, err := experimentDB.Prepare("CREATE TABLE IF NOT EXISTS experiments (id INTEGER PRIMARY KEY, slug TEXT, name TEXT, email TEXT, latitude REAL, longitude REAL, privatekey BLOB, nodesalt TEXT, Database TEXT)")
	if err != nil {
		panic(err)
	}
	statement.Exec()
}

func NewExperiment(jsonData []byte) (e Experiment, err error) {
	e = Experiment{}
	err = json.Unmarshal(jsonData, &e)
	fmt.Println(e.Name)
	//GET FROM DB
	var count int
	experimentDB.QueryRow("SELECT count(*) FROM experiments WHERE name=$1", e.Name).Scan(&count)
	if count == 0 {
		//INSERT
	} else {
		//GET
	}
	fmt.Println(count)

	e.Slug = strings.Replace(e.Name, " ", "-", -1)
	//crypto: https://medium.com/@raul_11817/golang-cryptography-rsa-asymmetric-algorithm-e91363a2f7b3
	privkey, err := rsa.GenerateKey(rand.Reader, 1024)
	e.PrivateKey = *privkey

	fileName := e.Slug + ".sqlite.db"
	if _, err = os.Stat(filepath.Join(DBExperimentsDir, fileName)); os.IsNotExist(err) {
		var dbfile *os.File
		dbfile, err = os.OpenFile(filepath.Join(DBExperimentsDir, fileName), os.O_RDONLY|os.O_CREATE, 0666)
		if err != nil {
			return
		}
		e.Database = *dbfile
	}

	db, err := sql.Open("sqlite3", filepath.Join(DBExperimentsDir, fileName))

	//Setting up table for nodes
	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS nodes (id INTEGER PRIMARY KEY, description TEXT, location TEXT, latitude REAL, longitude REAL, publickey BLOB)")

	if err != nil {
		return
	}
	statement.Exec()

	//Setting up table for readings
	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS readings (id INTEGER PRIMARY KEY, nodeID INT, deviceID TEXT, signal INT, Timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)")

	if err != nil {
		return
	}
	statement.Exec()

	//we also need to generate a config file and a private keys
	log.Printf("file: %v\n", e.Database)
	return
}

//from: https://snippets.aktagon.com/snippets/756-checking-if-a-row-exists-in-go-database-sql-and-sqlx-
