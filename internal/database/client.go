package database

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"database/sql"

	_ "modernc.org/sqlite"
)

var DB *sql.DB
var err error
var (
	pwd, _ = os.Getwd()
)

type Timeframe struct {
	ID       string `json:"id"`
	Year     int    `json:"year"`
	Month    int    `json:"month"`
	Day      int    `json:"day"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Duration string `json:"duration"`
	Project  string `json:"project"`
}

type Config struct {
	DailyHours time.Duration
}

func Connect() {
	DB, err = sql.Open("sqlite", pwd+"/timetrack.sqlite")
	if err != nil {
		log.Fatal(err)
		panic("Cannot connect to DB")
	}
	log.Println("Connected to Database...")
}

func checkIfEntryExists(entryID string) bool {
	return true
}

func CreateEntry(w http.ResponseWriter, r *http.Request) {

}

func GetEntryByID(w http.ResponseWriter, r *http.Request) {

}

func GetEntries(w http.ResponseWriter, r *http.Request) {
	var timeframes []Timeframe = []Timeframe{}
	var timef Timeframe

	rows, _ := DB.Query("SELECT * FROM timeframes")

	for rows.Next() {
		timef = Timeframe{}
		rows.Scan(&timef.ID, &timef.Year, &timef.Month, &timef.Day,
			&timef.Start, &timef.End, &timef.Duration, &timef.Project)
		timeframes = append(timeframes, timef)
		log.Printf("ID: %s", timef.ID)
	}

	err = json.NewEncoder(w).Encode(&timeframes)
	if err != nil {
		fmt.Print(err)
	}
}

func GetVersion() {
	var version string
	err = DB.QueryRow("SELECT SQLITE_VERSION()").Scan(&version)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(version)
}

func UpdateEntry(w http.ResponseWriter, r *http.Request) {

}

func DeleteEntry(w http.ResponseWriter, r *http.Request) {

}
