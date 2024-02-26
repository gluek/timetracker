// Copyright 2024 Gerrit Lükens. All rights reserved.
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB
var err error
var (
	pwd, _ = os.Getwd()
)

type Timeframe struct {
	ID        int    `json:"id"`
	Date      string `json:"date"`
	Year      int    `json:"year"`
	Month     int    `json:"month"`
	Day       int    `json:"day"`
	Start     string `json:"start"`
	End       string `json:"end"`
	Duration  string `json:"duration"`
	ProjectID string `json:"projectid"`
}

type Project struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Activity string `json:"activity"`
	Details  string `json:"details"`
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

	tableVars := "(id int, date string, year int, month int, day int, start string, end string, duration string, projectid string)"
	statement, err := DB.Prepare("CREATE TABLE IF NOT EXISTS timeframes " + tableVars)
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Created timesframes Table...")

	tableVars = "(id int, name string, activity string, details string)"
	statement, err = DB.Prepare("CREATE TABLE IF NOT EXISTS projects " + tableVars)
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Created projects Table...")

	_, err = GetProjectByID(0)
	if err != nil {
		statement, err = DB.Prepare("INSERT INTO projects (id, name, activity, details) VALUES (?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defaultProject := Project{ID: 0, Name: "NotAssigned", Activity: "", Details: ""}
		_, err = statement.Exec(defaultProject.ID, defaultProject.Name, defaultProject.Activity, defaultProject.Details)
		if err != nil {
			log.Fatal(err)
		}
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

func CreateRecord(timefr Timeframe) error {
	statement, err := DB.Prepare("INSERT INTO timeframes " +
		"(id, date, year, month, day, start, end, duration, projectid) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec(timefr.ID, timefr.Date, timefr.Year, timefr.Month, timefr.Day, timefr.Start, timefr.End, timefr.Duration, timefr.ProjectID)
	if err != nil {
		return err
	}
	return nil
}

func GetRecordByID(id int) error {
	var timefr Timeframe = Timeframe{}

	statement, err := DB.Prepare("SELECT * FROM timeframes WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}
	err = statement.QueryRow(id).Scan(&timefr.ID, &timefr.Date, &timefr.Year, &timefr.Month, &timefr.Day,
		&timefr.Start, &timefr.End, &timefr.Duration, &timefr.ProjectID)
	if err != nil {
		return err
	}
	return nil
}

func GetRecords() []Timeframe {
	var timeframes []Timeframe = []Timeframe{}
	var timefr Timeframe

	statement, err := DB.Prepare("SELECT * FROM timeframes")
	if err != nil {
		log.Fatal(err)
	}
	rows, _ := statement.Query()

	for rows.Next() {
		timefr = Timeframe{}
		rows.Scan(&timefr.ID, &timefr.Date, &timefr.Year, &timefr.Month, &timefr.Day,
			&timefr.Start, &timefr.End, &timefr.Duration, &timefr.ProjectID)
		timeframes = append(timeframes, timefr)
	}
	return timeframes
}

func GetRecordsMaxID() int {
	var maxID int
	statement, err := DB.Prepare("SELECT MAX(id) FROM timeframes")
	if err != nil {
		log.Fatal(err)
	}
	row := statement.QueryRow()
	if err != nil {
		log.Fatal(err)
	}
	row.Scan(&maxID)

	return maxID
}

func UpdateRecord(timefr Timeframe) error {

	statement, err := DB.Prepare("UPDATE timeframes SET " +
		"date=?, year=?, month=?, day=?, start=?, end=?, duration=?, projectid=? WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec(timefr.Date, timefr.Year, timefr.Month, timefr.Day, timefr.Start, timefr.End, timefr.Duration, timefr.ProjectID, timefr.ID)
	if err != nil {
		return err
	}
	return nil
}

func DeleteRecord(id int) error {
	statement, err := DB.Prepare("DELETE FROM timeframes WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec(id)
	if err != nil {
		return err
	}
	return nil
}

func CreateProject(project Project) error {
	statement, err := DB.Prepare("INSERT INTO projects " +
		"(id, name, activity, details) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec(project.ID, project.Name, project.Activity, project.Details)
	if err != nil {
		return err
	}
	return nil
}

func GetProjectByID(id int) (Project, error) {
	var project Project = Project{}

	statement, err := DB.Prepare("SELECT * FROM projects WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}
	err = statement.QueryRow(id).Scan(&project.ID, &project.Name, &project.Activity, &project.Details)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

func GetProjects() []Project {
	var projects []Project = []Project{}
	var project Project

	statement, err := DB.Prepare("SELECT * FROM projects")
	if err != nil {
		log.Fatal(err)
	}
	rows, _ := statement.Query()

	for rows.Next() {
		project = Project{}
		rows.Scan(&project.ID, &project.Name, &project.Activity, &project.Details)
		projects = append(projects, project)
	}
	return projects
}

func GetProjectsMaxID() int {
	var maxID int
	statement, err := DB.Prepare("SELECT MAX(id) FROM projects")
	if err != nil {
		log.Fatal(err)
	}
	row := statement.QueryRow()
	row.Scan(&maxID)

	return maxID
}

func UpdateProject(project Project) error {

	statement, err := DB.Prepare("UPDATE projects SET " +
		"name=?, activity=?, details=? WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec(project.Name, project.Activity, project.Details, project.ID)
	if err != nil {
		return err
	}
	return nil
}

func DeleteProject(id int) error {
	statement, err := DB.Prepare("DELETE FROM projects WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec(id)
	if err != nil {
		return err
	}
	return nil
}

func GetVersion() {
	var version string
	err = DB.QueryRow("SELECT SQLITE_VERSION()").Scan(&version)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(version)
}
