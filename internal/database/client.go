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
var dbVersion int = 2
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

type ProjectHoursLocationsDaily struct {
	Date      string   `json:"date" csv:"date"`
	Hours     string   `json:"workhours" csv:"Hours"`
	Projects  []string `json:"projects" csv:"Projects"`
	Locations []string `json:"locations" csv:"Workplace"`
}

type Location struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type LocationDays struct {
	Location
	Days int `json:"days"`
}

type Config struct {
	DailyHours time.Duration
}

func Connect() {
	DB, err = sql.Open("sqlite", pwd+"/timetrack.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to Database")
	DB.SetMaxOpenConns(2)

	// Check if new database
	var tableCount int
	statement, err := DB.Prepare("SELECT COUNT(*) FROM sqlite_master AS tables WHERE type='table'")
	if err != nil {
		log.Fatal(err)
	}
	err = statement.QueryRow().Scan(&tableCount)
	if err != nil {
		log.Fatal(err)
	}
	statement.Close()

	if tableCount == 0 {
		log.Println("New table, set db version")
		setDBVersion(dbVersion)
	}

	tableVars := `(
		id INTEGER PRIMARY KEY NOT NULL,
		date STRING NOT NULL,
		year INTEGER NOT NULL,
		month INTEGER NOT NULL,
		day INTEGER NOT NULL,
		start STRING NOT NULL,
		end STRING NOT NULL,
		duration STRING,
		projectid INTEGER NOT NULL,
		locationid INTEGER NOT NULL
	)`

	statement, err = DB.Prepare("CREATE TABLE IF NOT EXISTS timeframes " + tableVars)
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Created timeframes Table if necessary")
	statement.Close()

	tableVars = `(
		id INTEGER PRIMARY KEY NOT NULL,
		name STRING NOT NULL,
		activity STRING NOT NULL,
		details STRING NOT NULL
	)`
	statement, err = DB.Prepare("CREATE TABLE IF NOT EXISTS projects " + tableVars)
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Created projects Table if necessary")
	statement.Close()

	tableVars = `(
		id INTEGER PRIMARY KEY NOT NULL,
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
	log.Println("Created workplaces Table if necessary")
	statement.Close()

	if getDBVersion() != dbVersion {
		Migrations()
	}

	if len(GetProjects()) == 0 {
		log.Println("Created default projects")
		defaultProjects := []Project{
			{ID: 1, Name: "NotAssigned", Activity: "", Details: ""},
			{ID: 2, Name: "Vacation", Activity: "Vacation", Details: "Vacation"},
			{ID: 3, Name: "Sick", Activity: "Sick Days", Details: "Sick Days"},
			{ID: 4, Name: "Parental Leave", Activity: "Parental Leave", Details: "Parental Leave"},
		}
		for _, v := range defaultProjects {
			err = CreateProject(v)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if len(GetLocations()) == 0 {
		log.Println("Created default workplaces")
		defaultLocations := []Location{
			{ID: 1, Name: "Company"},
			{ID: 2, Name: "Home"},
			{ID: 3, Name: "Mobile"},
			{ID: 4, Name: "Trip"},
		}
		for _, v := range defaultLocations {
			err = CreateLocation(v)
			if err != nil {
				log.Fatal(err)
			}
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

func CreateRecord(timefr Timeframe) error {
	statement, err := DB.Prepare("INSERT INTO timeframes " +
		"(date, year, month, day, start, end, duration, projectid, locationid) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(timefr.Date, timefr.Year, timefr.Month, timefr.Day,
		timefr.Start, timefr.End, timefr.Duration, timefr.ProjectID, timefr.LocationID)
	if err != nil {
		return err
	}
	return nil
}

func GetRecordByID(id int) error {
	var timefr Timeframe = Timeframe{}

	statement, err := DB.Prepare("SELECT * FROM timeframes WHERE id=?")
	if err != nil {
		return err
	}
	defer statement.Close()

	err = statement.QueryRow(id).Scan(&timefr.ID, &timefr.Date, &timefr.Year, &timefr.Month, &timefr.Day,
		&timefr.Start, &timefr.End, &timefr.Duration, &timefr.ProjectID, &timefr.LocationID)
	if err != nil {
		return err
	}
	return nil
}

func getTimeframes(sqlString string, args ...any) ([]Timeframe, error) {
	var timeframes []Timeframe = []Timeframe{}
	var timefr Timeframe

	statement, err := DB.Prepare(sqlString)
	if err != nil {
		return []Timeframe{}, err
	}
	defer statement.Close()

	rows, err := statement.Query(args...)
	if err != nil {
		return []Timeframe{}, err
	}
	defer rows.Close()

	for rows.Next() {
		timefr = Timeframe{}
		err = rows.Scan(&timefr.ID, &timefr.Date, &timefr.Year, &timefr.Month, &timefr.Day,
			&timefr.Start, &timefr.End, &timefr.Duration, &timefr.ProjectID, &timefr.LocationID)
		if err != nil {
			return []Timeframe{}, err
		}

		timeframes = append(timeframes, timefr)
	}
	return timeframes, nil
}

func GetRecords() []Timeframe {
	statement := "SELECT * FROM timeframes"
	timeframes, err := getTimeframes(statement)
	if err != nil {
		log.Printf("error GetRecords: %v", err)
		return []Timeframe{}
	}
	return timeframes
}

func GetRecordsForDate(date time.Time) []Timeframe {
	statement := "SELECT * FROM timeframes WHERE date=?;"
	timeframes, err := getTimeframes(statement, date.Format("2006-01-02"))
	if err != nil {
		log.Printf("error GetRecordsForDate: %v", err)
		return []Timeframe{}
	}
	return timeframes
}

func GetRecordsForProjectAndMonth(year int, month int, projectid int) []Timeframe {
	statement := "SELECT * FROM timeframes WHERE year=? AND month=? AND projectid=?"
	timeframes, err := getTimeframes(statement, year, month, projectid)
	if err != nil {
		log.Printf("error GetRecordsForProjectAndMonth: %v", err)
		return []Timeframe{}
	}
	return timeframes
}

func GetRecordsForProjectAndYear(year time.Time, projectid int) []Timeframe {
	statement := "SELECT * FROM timeframes WHERE year=? AND projectid=?;"
	timeframes, err := getTimeframes(statement, year.Year(), projectid)
	if err != nil {
		log.Printf("error GetRecordsForProjectAndYear: %v", err)
		return []Timeframe{}
	}
	return timeframes
}

func GetRecordsForProjectAndYearUntilToday(year time.Time, day time.Time, projectid int) []Timeframe {
	var endDate time.Time
	if year.Year() < time.Now().Year() {
		endDate, err = time.Parse("2006-01-02", fmt.Sprintf("%d-12-31", year.Year()))
		if err != nil {
			log.Printf("error GetRecordsForProjectAndYearUntilToday: %v", err)
			return []Timeframe{}
		}
	} else {
		endDate = day
	}

	statement := "SELECT * FROM timeframes WHERE year=? AND date<=date(?) AND projectid=?;"
	timeframes, err := getTimeframes(statement, year.Year(), endDate.Format("2006-01-02"), projectid)
	if err != nil {
		log.Printf("error GetRecordsForProjectAndYearUntilToday: %v", err)
		return []Timeframe{}
	}
	return timeframes
}

func UpdateRecord(timefr Timeframe) error {
	statement, err := DB.Prepare("UPDATE timeframes SET " +
		"date=?, year=?, month=?, day=?, start=?, end=?, duration=?, projectid=?, locationid=? WHERE id=?")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(timefr.Date, timefr.Year, timefr.Month, timefr.Day,
		timefr.Start, timefr.End, timefr.Duration, timefr.ProjectID, timefr.LocationID, timefr.ID)
	if err != nil {
		return err
	}
	return nil
}

func DeleteRecord(id int) error {
	statement, err := DB.Prepare("DELETE FROM timeframes WHERE id=?")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(id)
	if err != nil {
		return err
	}
	return nil
}

func CreateProject(project Project) error {
	statement, err := DB.Prepare("INSERT INTO projects " +
		"(name, activity, details) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(project.Name, project.Activity, project.Details)
	if err != nil {
		return err
	}
	return nil
}

func GetProjectByID(id int) (Project, error) {
	var project Project = Project{}

	statement, err := DB.Prepare("SELECT * FROM projects WHERE id=?")
	if err != nil {
		return Project{}, err
	}

	row := statement.QueryRow(id)
	err = row.Scan(&project.ID, &project.Name, &project.Activity, &project.Details)
	if err != nil {
		return Project{}, err
	}
	statement.Close()
	return project, nil
}

func GetProjects() []Project {
	var projects []Project = []Project{}
	var project Project

	statement, err := DB.Prepare("SELECT * FROM projects")
	if err != nil {
		log.Printf("could not prepare statement GetProjects: %v", err)
		return []Project{}
	}
	defer statement.Close()

	rows, err := statement.Query()
	if err != nil {
		log.Printf("could not query database GetProjects: %v", err)
		return []Project{}
	}
	defer rows.Close()

	for rows.Next() {
		project = Project{}
		err = rows.Scan(&project.ID, &project.Name, &project.Activity, &project.Details)
		if err != nil {
			log.Printf("could not scan rows GetProjects: %v", err)
			return []Project{}
		}
		projects = append(projects, project)
	}
	return projects
}

func UpdateProject(project Project) error {
	statement, err := DB.Prepare("UPDATE projects SET " +
		"name=?, activity=?, details=? WHERE id=?")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(project.Name, project.Activity, project.Details, project.ID)
	if err != nil {
		return err
	}
	return nil
}

func DeleteProject(id int) error {
	statement, err := DB.Prepare("DELETE FROM projects WHERE id=?")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(id)
	if err != nil {
		return err
	}
	return nil
}

func GetProjectsForDate(date time.Time) map[int]string {
	statement, err := DB.Prepare("SELECT projectid FROM timeframes WHERE date=?")
	if err != nil {
		log.Printf("could not prepare statement GetProjectsForDate: %v", err)
		return map[int]string{}
	}

	rows, err := statement.Query(date.Format("2006-01-02"))
	if err != nil {
		log.Printf("could not query database GetProjectsForDate: %v", err)
		return map[int]string{}
	}
	defer rows.Close()

	statement.Close()

	projectids := map[int]string{}
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			log.Printf("could not scan row GetProjectsForDate: %v", err)
			return map[int]string{}
		}
		projectName, err := GetProjectByID(id)
		if err != nil {
			log.Printf("could not get project for idx %d GetProjectsForDate: %v", id, err)
			return map[int]string{}
		}
		projectids[id] = projectName.Name
	}
	return projectids
}

func CreateLocation(location Location) error {
	statement, err := DB.Prepare("INSERT INTO workplaces " +
		"(location) VALUES (?)")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(location.Name)
	if err != nil {
		return err
	}
	return nil
}

func GetLocations() []Location {
	statement, err := DB.Prepare("SELECT * FROM workplaces;")
	if err != nil {
		log.Printf("could not prepare statement GetLocations: %v", err)
		return []Location{}
	}
	defer statement.Close()

	rows, err := statement.Query()
	if err != nil {
		log.Printf("could not query database GetLocations: %v", err)
		return []Location{}
	}
	defer rows.Close()

	locations := []Location{}
	for rows.Next() {
		var location Location
		err = rows.Scan(&location.ID, &location.Name)
		if err != nil {
			log.Printf("could not scan rows GetLocations: %v", err)
			return []Location{}
		}
		locations = append(locations, location)
	}
	return locations
}
func GetLocationByID(id int) (Location, error) {
	var location Location = Location{}

	statement, err := DB.Prepare("SELECT * FROM workplaces WHERE id=?")
	if err != nil {
		return Location{}, err
	}
	defer statement.Close()

	err = statement.QueryRow(id).Scan(&location.ID, &location.Name)
	if err != nil {
		return Location{}, err
	}
	return location, nil
}

func GetLocationsForDate(date time.Time) map[int]string {
	statement, err := DB.Prepare("SELECT locationid FROM timeframes WHERE date=?")
	if err != nil {
		log.Printf("could not prepare statement GetLocationsForDate: %v", err)
		return map[int]string{}
	}
	defer statement.Close()

	rows, err := statement.Query(date.Format("2006-01-02"))
	if err != nil {
		log.Printf("could not query database GetLocationsForDate: %v", err)
		return map[int]string{}
	}
	defer rows.Close()

	locationids := map[int]string{}
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			log.Printf("could not scan rows GetLocationsForDate: %v", err)
			return map[int]string{}
		}
		projectName, err := GetLocationByID(id)
		if err != nil {
			log.Printf("could not get location name for idx %d GetLocationsForDate: %v", id, err)
			return map[int]string{}
		}
		locationids[id] = projectName.Name
	}
	return locationids
}

func GetLocationDaysForMonth(month time.Time) []LocationDays {
	statement, err := DB.Prepare(`
		SELECT workplaces.id, workplaces.location, COUNT(DISTINCT date)
		FROM timeframes INNER JOIN workplaces 
		ON timeframes.locationid = workplaces.id
		WHERE timeframes.year=? AND timeframes.month=? AND timeframes.projectid NOT IN (2,3,4)
		GROUP BY workplaces.location;`)
	if err != nil {
		log.Printf("could not prepare statement GetLocationDaysForMonth: %v", err)
		return []LocationDays{}
	}
	defer statement.Close()

	rows, err := statement.Query(month.Year(), int(month.Month()))
	if err != nil {
		log.Printf("could not query database GetLocationDaysForMonth: %v", err)
		return []LocationDays{}
	}
	defer rows.Close()

	locationDays := []LocationDays{}
	for rows.Next() {
		location := LocationDays{Location{ID: 0, Name: ""}, 0}
		err = rows.Scan(&location.Location.ID, &location.Location.Name, &location.Days)
		if err != nil {
			log.Printf("could not scan rows GetLocationDaysForMonth: %v", err)
			return []LocationDays{}
		}
		locationDays = append(locationDays, location)
	}
	return locationDays
}

func GetLocationDaysForYear(year time.Time) []LocationDays {
	statement, err := DB.Prepare(`
		SELECT workplaces.id, workplaces.location, COUNT(DISTINCT date)
		FROM timeframes INNER JOIN workplaces 
		ON timeframes.locationid = workplaces.id
		WHERE timeframes.year=? AND timeframes.projectid NOT IN (2,3,4)
		GROUP BY workplaces.location;`)
	if err != nil {
		log.Printf("could not prepare statement GetLocationDaysForYear: %v", err)
		return []LocationDays{}
	}
	defer statement.Close()

	rows, err := statement.Query(year.Year())
	if err != nil {
		log.Printf("could not query database GetLocationDaysForYear: %v", err)
		return []LocationDays{}
	}
	defer rows.Close()

	locationDays := []LocationDays{}
	for rows.Next() {
		location := LocationDays{Location{ID: 0, Name: ""}, 0}
		err = rows.Scan(&location.Location.ID, &location.Location.Name, &location.Days)
		if err != nil {
			log.Printf("could not scan rows GetLocationDaysForYear: %v", err)
			return []LocationDays{}
		}
		locationDays = append(locationDays, location)
	}
	return locationDays
}

func GetVersion() {
	var version string
	err = DB.QueryRow("SELECT SQLITE_VERSION()").Scan(&version)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(version)
}
