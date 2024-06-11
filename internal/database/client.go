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
	ID         int    `json:"id"`
	Date       string `json:"date"`
	Year       int    `json:"year"`
	Month      int    `json:"month"`
	Day        int    `json:"day"`
	Start      string `json:"start"`
	End        string `json:"end"`
	Duration   string `json:"duration"`
	ProjectID  int    `json:"projectid"`
	LocationID int    `json:"locationid"`
}

type Project struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Activity string `json:"activity"`
	Details  string `json:"details"`
}

type ProjectHours struct {
	Hours string `json:"workhours"`
	Project
}

type ProjectHoursDaily struct {
	Date     string   `json:"date"`
	Hours    string   `json:"workhours"`
	Projects []string `json:"projects"`
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

	tableVars := `(
		id int PRIMARY KEY,
		date string NOT NULL,
		year int NOT NULL,
		month int NOT NULL,
		day int NOT NULL,
		start string NOT NULL,
		end string NOT NULL,
		duration string,
		projectid int,
		locationid int
	)`

	statement, err := DB.Prepare("CREATE TABLE IF NOT EXISTS timeframes " + tableVars)
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Created timeframes Table...")

	tableVars = `(
		id int PRIMARY KEY,
		name string NOT NULL,
		activity string NOT NULL,
		details string NOT NULL
	)`
	statement, err = DB.Prepare("CREATE TABLE IF NOT EXISTS projects " + tableVars)
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Created projects Table...")

	tableVars = `(
		id INT PRIMARY KEY,
		location STRING NOT NULL
	)`
	statement, err = DB.Prepare("CREATE TABLE IF NOT EXISTS workplaces " + tableVars)
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Created workplaces Table...")

	Migrations()

	_, err = GetProjectByID(0)
	if err != nil {
		log.Println("Created default projects...")
		defaultProject := Project{ID: 0, Name: "NotAssigned", Activity: "", Details: ""}
		vacationProject := Project{ID: 1, Name: "Vacation", Activity: "Vacation", Details: "Vacation"}
		sickdaysProject := Project{ID: 2, Name: "Sick", Activity: "Sick Days", Details: "Sick Days"}
		parentalleaveProject := Project{ID: 3, Name: "Parental Leave", Activity: "Parental Leave", Details: "Parental Leave"}
		err = CreateProject(defaultProject)
		if err != nil {
			log.Fatal(err)
		}
		err = CreateProject(vacationProject)
		if err != nil {
			log.Fatal(err)
		}
		err = CreateProject(sickdaysProject)
		if err != nil {
			log.Fatal(err)
		}
		err = CreateProject(parentalleaveProject)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func Close() {
	err = DB.Close()
	if err != nil {
		log.Println("Could not close database")
		log.Fatal(err)
	}
	log.Println("Database closed")
}

func Migrations() {
	version := getDBVersion()

	switch version {
	case 0:
		statement, err := DB.Prepare("ALTER TABLE timeframes ADD COLUMN locationid int;")
		if err != nil {
			log.Fatal(err)
		}
		_, err = statement.Exec()
		if err != nil {
			log.Fatal(err)
		}
		setDBVersion(1)
		log.Println("Migration to Version 1")
		fallthrough
	default:
		log.Printf("DB Version: %d", getDBVersion())
	}
}

func getDBVersion() int {
	var version int
	statement, err := DB.Prepare("PRAGMA user_version;")
	if err != nil {
		log.Fatal(err)
	}
	err = statement.QueryRow().Scan(&version)
	if err != nil {
		log.Fatal(err)
	}
	return version
}

func setDBVersion(version int) {
	statement, err := DB.Prepare(fmt.Sprintf("PRAGMA user_version = %d;", version))
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
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
	rows, err := statement.Query()
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		timefr = Timeframe{}
		err = rows.Scan(&timefr.ID, &timefr.Date, &timefr.Year, &timefr.Month, &timefr.Day,
			&timefr.Start, &timefr.End, &timefr.Duration, &timefr.ProjectID)
		if err != nil {
			log.Fatal(err)
		}

		timeframes = append(timeframes, timefr)
	}
	return timeframes
}

func GetRecordsForDate(date time.Time) []Timeframe {
	var timeframes []Timeframe = []Timeframe{}
	var timefr Timeframe

	statement, err := DB.Prepare("SELECT * FROM timeframes WHERE date=?")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := statement.Query(date.Format("2006-01-02"))
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		timefr = Timeframe{}
		err = rows.Scan(&timefr.ID, &timefr.Date, &timefr.Year, &timefr.Month, &timefr.Day,
			&timefr.Start, &timefr.End, &timefr.Duration, &timefr.ProjectID)
		if err != nil {
			log.Fatal(err)
		}
		timeframes = append(timeframes, timefr)
	}
	return timeframes
}

func GetRecordsForProjectAndMonth(year int, month int, projectid int) []Timeframe {
	var timeframes []Timeframe = []Timeframe{}
	var timefr Timeframe

	statement, err := DB.Prepare("SELECT * FROM timeframes WHERE year=? AND month=? AND projectid=?")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := statement.Query(year, month, projectid)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		timefr = Timeframe{}
		err = rows.Scan(&timefr.ID, &timefr.Date, &timefr.Year, &timefr.Month, &timefr.Day,
			&timefr.Start, &timefr.End, &timefr.Duration, &timefr.ProjectID)
		if err != nil {
			log.Fatal(err)
		}
		timeframes = append(timeframes, timefr)
	}
	return timeframes
}

func GetRecordsForProjectAndYear(year time.Time, projectid int) []Timeframe {
	var timeframes []Timeframe = []Timeframe{}
	var timefr Timeframe

	statement, err := DB.Prepare("SELECT * FROM timeframes WHERE year=? AND projectid=?")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := statement.Query(year.Year(), projectid)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		timefr = Timeframe{}
		err = rows.Scan(&timefr.ID, &timefr.Date, &timefr.Year, &timefr.Month, &timefr.Day,
			&timefr.Start, &timefr.End, &timefr.Duration, &timefr.ProjectID)
		if err != nil {
			log.Fatal(err)
		}
		timeframes = append(timeframes, timefr)
	}
	return timeframes
}

func GetRecordsForProjectAndYearUntilToday(year time.Time, day time.Time, projectid int) []Timeframe {
	var timeframes []Timeframe = []Timeframe{}
	var timefr Timeframe

	var endDate time.Time
	if year.Year() < time.Now().Year() {
		endDate, err = time.Parse("2006-01-02", fmt.Sprintf("%d-12-31", year.Year()))
		if err != nil {
			log.Println(err)
		}
	} else {
		endDate = day
	}

	statement, err := DB.Prepare("SELECT * FROM timeframes WHERE year=? AND date<=date(?) AND projectid=?")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := statement.Query(year.Year(), endDate.Format("2006-01-02"), projectid)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		timefr = Timeframe{}
		err = rows.Scan(&timefr.ID, &timefr.Date, &timefr.Year, &timefr.Month, &timefr.Day,
			&timefr.Start, &timefr.End, &timefr.Duration, &timefr.ProjectID)
		if err != nil {
			log.Fatal(err)
		}
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
	err = row.Scan(&maxID)
	if err != nil {
		log.Fatal(err)
	}

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
	rows, err := statement.Query()
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		project = Project{}
		err = rows.Scan(&project.ID, &project.Name, &project.Activity, &project.Details)
		if err != nil {
			log.Fatal(err)
		}
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
	err = row.Scan(&maxID)
	if err != nil {
		log.Fatal(err)
	}

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

func GetProjectsForDate(date time.Time) map[int]string {
	statement, err := DB.Prepare("SELECT projectid FROM timeframes WHERE date=?")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := statement.Query(date.Format("2006-01-02"))
	if err != nil {
		log.Fatal(err)
	}
	projectids := map[int]string{}
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		projectName, err := GetProjectByID(id)
		if err != nil {
			log.Print(err)
		}
		projectids[id] = projectName.Name
	}
	return projectids
}

func GetVersion() {
	var version string
	err = DB.QueryRow("SELECT SQLITE_VERSION()").Scan(&version)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(version)
}
