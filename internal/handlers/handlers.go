package handlers

import (
	"fmt"
	"local/timetracker/internal/components"
	"local/timetracker/internal/database"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

var (
	err         error
	tfList      []database.Timeframe
	projectList = []database.Project{{
		ID:       0,
		Name:     "NotAssigned",
		Activity: "",
		Details:  "",
	}}
	globalID        int
	globalIDProject int    = 1
	activeDate      string = time.Now().Format("2006-01-02")
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	component := components.HomePage(
		database.GetRecordsForDate(activeDate),
		database.GetProjects(),
		activeDate,
		workTotalByDate(activeDate).String(),
		workDeltaWeek(workTotalWeek(activeDate)).String())
	component.Render(r.Context(), w)
}

func RecordsPageHandler(w http.ResponseWriter, r *http.Request) {
	component := components.Records(
		database.GetRecordsForDate(activeDate),
		database.GetProjects(),
		activeDate,
		workTotalByDate(activeDate).String(),
		workDeltaWeek(workTotalWeek(activeDate)).String())
	component.Render(r.Context(), w)
}

func RecordsHandler(w http.ResponseWriter, r *http.Request) {
	component := components.RecordList(
		database.GetRecordsForDate(activeDate),
		database.GetProjects())
	component.Render(r.Context(), w)
}

func ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	components.Projects(database.GetProjects()).Render(r.Context(), w)
}

func SummaryHandler(w http.ResponseWriter, r *http.Request) {
	components.MonthlySummary(GetProjectHours()).Render(r.Context(), w)
}

func atoi(s string) int {
	value, err := strconv.Atoi(s)
	if err != nil {
		log.Println(err)
	}
	return value
}

func parseDate(date string) (int, int, int) {

	dateTime, _ := time.Parse("2006-01-02", date)
	year, month, day := dateTime.Date()
	monthInt := int(month)
	return year, monthInt, day
}

func findIDTimeframe(id int) int {
	for index, timeframe := range tfList {
		if timeframe.ID == id {
			return index
		}
	}
	return -1
}

func findIDProject(id int) int {
	for index, project := range projectList {
		if project.ID == id {
			return index
		}
	}
	return -1
}

func workTotalByDate(date string) time.Duration {
	timeTotal, err := time.ParseDuration("0s")
	if err != nil {
		log.Print(err)
	}
	records := database.GetRecordsForDate(date)
	for _, record := range records {
		timeStart, err := time.Parse("15:04", record.Start)
		if err != nil {
			log.Print(err)
		}
		timeEnd, err := time.Parse("15:04", record.End)
		if err != nil {
			log.Print(err)
		}
		diffTime := timeEnd.Sub(timeStart)
		timeTotal += diffTime
	}
	return timeTotal
}

func workTotalForRecords(timeframes []database.Timeframe) time.Duration {
	timeTotal, err := time.ParseDuration("0s")
	if err != nil {
		log.Print(err)
	}
	for _, record := range timeframes {
		timeStart, err := time.Parse("15:04", record.Start)
		if err != nil {
			log.Print(err)
		}
		timeEnd, err := time.Parse("15:04", record.End)
		if err != nil {
			log.Print(err)
		}
		diffTime := timeEnd.Sub(timeStart)
		timeTotal += diffTime
	}
	return timeTotal
}

func weekDaysByDate(date string) []time.Time {
	var weekDays []time.Time
	daysInAWeek := []int{1, 2, 3, 4, 5, 6, 7}

	dateTime, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Print(err)
	}
	for _, weekday := range daysInAWeek {
		dayOfWeek := int(dateTime.Weekday())
		if dayOfWeek == 0 {
			dayOfWeek = 7
		}
		weekdayOffset := weekday - dayOfWeek
		weekDays = append(weekDays, dateTime.AddDate(0, 0, weekdayOffset))
	}
	return weekDays
}

func workTotalWeek(date string) time.Duration {
	workTotalDuration, err := time.ParseDuration("0s")
	if err != nil {
		log.Print(err)
	}
	var daysInWeek []time.Time = weekDaysByDate(date)
	for _, day := range daysInWeek {
		dayDuration := workTotalByDate(day.Format("2006-01-02"))
		workTotalDuration += dayDuration
	}
	return workTotalDuration
}

func workDeltaWeek(workTotalDuration time.Duration) time.Duration {
	workTotalTarget, err := time.ParseDuration(viper.GetString("worktime_per_week"))
	workDelta := workTotalDuration - workTotalTarget
	if err != nil {
		log.Println(err)
	}
	return workDelta
}

func ChangeDate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	activeDate = r.FormValue("dateofrecord")
	fmt.Printf("Date changed to %s\n", activeDate)
	RecordsPageHandler(w, r)
}

func CreateRecord(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	year, month, day := parseDate(r.FormValue("dateofrecord"))
	var timeframe = database.Timeframe{
		ID:        database.GetRecordsMaxID() + 1,
		Date:      r.FormValue("dateofrecord"),
		Year:      year,
		Month:     month,
		Day:       day,
		Start:     r.FormValue("start"),
		End:       r.FormValue("end"),
		Duration:  "",
		ProjectID: atoi(r.FormValue("project")),
	}
	globalID += 1
	err = database.CreateRecord(timeframe)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created Record %s %s ID: %d\n", timeframe.Start, timeframe.End, timeframe.ProjectID)
	RecordsPageHandler(w, r)
}

func UpdateRecord(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	year, month, day := parseDate(r.FormValue("dateofrecord"))
	id := r.PathValue("id")
	timefr := database.Timeframe{
		ID:        atoi(id),
		Date:      r.FormValue("dateofrecord"),
		Year:      year,
		Month:     month,
		Day:       day,
		Start:     r.FormValue("start"),
		End:       r.FormValue("end"),
		Duration:  "",
		ProjectID: atoi(r.FormValue("project")),
	}
	err = database.UpdateRecord(timefr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Update Record with ID %s\n", id)
	RecordsPageHandler(w, r)
}

func DeleteRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err = database.DeleteRecord(atoi(id))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Remove Record with ID: %v\n", id)
	RecordsPageHandler(w, r)
}

func CreateProject(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var project = database.Project{
		ID:       database.GetProjectsMaxID() + 1,
		Name:     r.FormValue("projectName"),
		Activity: r.FormValue("activity"),
		Details:  r.FormValue("details"),
	}
	globalIDProject += 1
	err = database.CreateProject(project)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Create Project %s %s %s\n", project.Activity, project.Details, project.Name)
	ProjectsHandler(w, r)
}

func UpdateProject(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.PathValue("id")
	project := database.Project{
		ID:       atoi(id),
		Name:     r.FormValue("projectName"),
		Activity: r.FormValue("activity"),
		Details:  r.FormValue("details"),
	}
	err = database.UpdateProject(project)
	fmt.Printf("Update Project with ID %s\n", id)
	ProjectsHandler(w, r)
}

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err = database.DeleteProject(atoi(id))
	fmt.Printf("Remove Project with ID: %v\n", id)
	ProjectsHandler(w, r)
}

func GetProjectHours() []database.ProjectHours {
	today := time.Now()
	projectsList := []database.ProjectHours{}
	projects := database.GetProjects()
	for _, project := range projects {
		records := database.GetRecordsForProjectAndTime(today.Year(), int(today.Month()), project.ID)
		duration := workTotalForRecords(records)
		projectHour := database.ProjectHours{
			Project: project,
			Hours:   fmt.Sprintf("%.2f", duration.Hours()),
		}
		projectsList = append(projectsList, projectHour)
	}
	return projectsList
}
