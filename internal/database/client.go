package database

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"database/sql"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

var DB *sql.DB
var err error
var (
	pwd, _ = os.Getwd()
)

type Timeframe struct {
	ID       string `json:"id"`
	Date     string `json:"date"`
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

	tableVars := "(id string, year int, month int, day int, start string, end string, duration string, project string)"
	statement, err := DB.Prepare("CREATE TABLE IF NOT EXISTS timeframes " + tableVars)
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
}

func Close() {
	err = DB.Close()
	if err != nil {
		log.Fatal(err)
		log.Println("Could not close database")
	}
	log.Println("Database closed")
}

func CreateEntry(w http.ResponseWriter, r *http.Request) {
	var timefr Timeframe
	json.NewDecoder(r.Body).Decode(&timefr)

	statement, err := DB.Prepare("INSERT INTO timeframes " +
		"(id, year, month, day, start, end, duration, project) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec(timefr.ID, timefr.Year, timefr.Month, timefr.Day, timefr.Start, timefr.End, timefr.Duration, timefr.Project)
	if err != nil {
		http.Error(w, "Failed to create timeframe", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Timeframe created succesfully")
}

func GetEntryByID(w http.ResponseWriter, r *http.Request) {
	var timefr Timeframe = Timeframe{}

	vars := mux.Vars(r)
	idStr := vars["id"]

	statement, err := DB.Prepare("SELECT * FROM timeframes WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}
	err = statement.QueryRow(idStr).Scan(&timefr.ID, &timefr.Year, &timefr.Month, &timefr.Day,
		&timefr.Start, &timefr.End, &timefr.Duration, &timefr.Project)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not find timeframe with id=%s", idStr), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(&timefr)
	if err != nil {
		fmt.Print(err)
	}
}

func GetEntries(w http.ResponseWriter, r *http.Request) {
	var timeframes []Timeframe = []Timeframe{}
	var timefr Timeframe

	statement, err := DB.Prepare("SELECT * FROM timeframes")
	if err != nil {
		log.Fatal(err)
	}
	rows, _ := statement.Query()

	for rows.Next() {
		timefr = Timeframe{}
		rows.Scan(&timefr.ID, &timefr.Year, &timefr.Month, &timefr.Day,
			&timefr.Start, &timefr.End, &timefr.Duration, &timefr.Project)
		timeframes = append(timeframes, timefr)
	}

	err = json.NewEncoder(w).Encode(&timeframes)
	if err != nil {
		fmt.Print(err)
	}
}

func UpdateEntry(w http.ResponseWriter, r *http.Request) {
	var timefr Timeframe
	json.NewDecoder(r.Body).Decode(&timefr)

	vars := mux.Vars(r)
	idStr := timefr.ID
	if vars["id"] != idStr {
		http.Error(w, "ID in request and provided data do not match", http.StatusBadRequest)
		return
	}

	statement, err := DB.Prepare("UPDATE timeframes SET " +
		"year=?, month=?, day=?, start=?, end=?, duration=?, project=? WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec(timefr.Year, timefr.Month, timefr.Day, timefr.Start, timefr.End, timefr.Duration, timefr.Project, timefr.ID)
	if err != nil {
		http.Error(w, "Failed to update timeframe", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Timeframe updated succesfully")
}

func DeleteEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	statement, err := DB.Prepare("DELETE FROM timeframes WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not delete timeframe with id=%s", idStr), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Timeframe deleted succesfully")
}

func GetVersion() {
	var version string
	err = DB.QueryRow("SELECT SQLITE_VERSION()").Scan(&version)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(version)
}
