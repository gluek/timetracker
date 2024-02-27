package handlers

import (
	"fmt"
	"local/timetracker/internal/components"
	"local/timetracker/internal/database"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	component := components.HomePage(database.GetRecordsForDate(activeDate), database.GetProjects(), activeDate)
	component.Render(r.Context(), w)
}

func RecordsPageHandler(w http.ResponseWriter, r *http.Request) {
	components.Records(database.GetRecordsForDate(activeDate), database.GetProjects(), activeDate).Render(r.Context(), w)
}

func RecordsHandler(w http.ResponseWriter, r *http.Request) {
	components.RecordList(database.GetRecordsForDate(activeDate), database.GetProjects()).Render(r.Context(), w)
}

func ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	components.Projects(database.GetProjects()).Render(r.Context(), w)
}

func atoi(s string) int {
	value, _ := strconv.Atoi(s)
	return value
}

func parseDate(date string) (int, int, int) {

	splitString := strings.Split(date, "-")
	year, month, day := splitString[0], splitString[1], splitString[2]
	return atoi(year), atoi(month), atoi(day)
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

func ChangeDate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	activeDate = r.FormValue("dateofrecord")
	fmt.Printf("Date changed to %s\n", activeDate)
	RecordsHandler(w, r)
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
		ProjectID: r.FormValue("project"),
	}
	globalID += 1
	err = database.CreateRecord(timeframe)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created Record %s %s %s\n", timeframe.Start, timeframe.End, timeframe.ProjectID)
	RecordsHandler(w, r)
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
		ProjectID: r.FormValue("project"),
	}
	err = database.UpdateRecord(timefr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Update Record with ID %s\n", id)
	RecordsHandler(w, r)
}

func DeleteRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err = database.DeleteRecord(atoi(id))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Remove Record with ID: %v\n", id)
	RecordsHandler(w, r)
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
