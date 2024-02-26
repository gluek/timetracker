package handlers

import (
	"fmt"
	"local/timetracker/internal/components"
	"local/timetracker/internal/database"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	err         error
	tfList      []database.Timeframe
	projectList = []database.Project{{
		ID:       "0",
		Name:     "NotAssigned",
		Activity: "",
		Details:  "",
	}}
	globalID        int
	globalIDProject int = 1
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	component := components.HomePage(tfList, projectList)
	component.Render(r.Context(), w)
}

func ProjectsPage(w http.ResponseWriter, r *http.Request) {
	component := components.ProjectsPage(projectList)
	component.Render(r.Context(), w)
}

func Root(w http.ResponseWriter, r *http.Request) {
	HomePage(w, r)
}

func recordsHandler(w http.ResponseWriter, r *http.Request) {
	components.Records(database.GetRecords(), projectList).Render(r.Context(), w)
}

func projectsHandler(w http.ResponseWriter, r *http.Request) {
	components.Projects(projectList).Render(r.Context(), w)
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

func findIDTimeframe(id string) int {
	for index, timeframe := range tfList {
		if timeframe.ID == id {
			return index
		}
	}
	return -1
}

func findIDProject(id string) int {
	for index, project := range projectList {
		if project.ID == id {
			return index
		}
	}
	return -1
}

func MockCreateRecord(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	year, month, day := parseDate(r.FormValue("dateofrecord"))
	var timeframe = database.Timeframe{
		ID:        strconv.Itoa(globalID),
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
	tfList = append(tfList, timeframe)
	fmt.Printf("%s %s %s Len of tfList: %d\n", timeframe.Start, timeframe.End, timeframe.ProjectID, len(tfList))
	recordsHandler(w, r)
}

func MockUpdateRecord(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	year, month, day := parseDate(r.FormValue("dateofrecord"))
	id := r.PathValue("id")
	index := findIDTimeframe(id)
	tfList[index] = database.Timeframe{
		ID:        id,
		Date:      r.FormValue("dateofrecord"),
		Year:      year,
		Month:     month,
		Day:       day,
		Start:     r.FormValue("start"),
		End:       r.FormValue("end"),
		Duration:  "",
		ProjectID: r.FormValue("project"),
	}
	fmt.Printf("Record with ID %s updated\n", id)
	recordsHandler(w, r)
}

func MockDeleteRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	index := findIDTimeframe(id)
	tfList = append(tfList[:index], tfList[index+1:]...)
	fmt.Printf("Remove Record with ID: %v\n", id)
	recordsHandler(w, r)
}

func MockCreateProject(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var project = database.Project{
		ID:       strconv.Itoa(globalIDProject),
		Name:     r.FormValue("projectName"),
		Activity: r.FormValue("activity"),
		Details:  r.FormValue("details"),
	}
	globalIDProject += 1
	projectList = append(projectList, project)
	fmt.Printf("%s %s %s Len of projectList: %d\n", project.Activity, project.Details, project.Name, len(projectList)-1)
	projectsHandler(w, r)
}

func MockUpdateProject(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.PathValue("id")
	index := findIDProject(id)
	projectList[index] = database.Project{
		ID:       strconv.Itoa(globalIDProject),
		Name:     r.FormValue("projectName"),
		Activity: r.FormValue("activity"),
		Details:  r.FormValue("details"),
	}
	fmt.Printf("Project with ID %s updated\n", id)
	projectsHandler(w, r)
}

func MockDeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	index := findIDProject(id)
	projectList = append(projectList[:index], projectList[index+1:]...)
	fmt.Printf("Remove Project with ID: %v\n", id)
	projectsHandler(w, r)
}

func CreateRecord(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	year, month, day := parseDate(r.FormValue("dateofrecord"))
	var timeframe = database.Timeframe{
		ID:        strconv.Itoa(globalID),
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
	recordsHandler(w, r)
}

func UpdateRecord(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	year, month, day := parseDate(r.FormValue("dateofrecord"))
	id := r.PathValue("id")
	timefr := database.Timeframe{
		ID:        id,
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
	recordsHandler(w, r)
}

func DeleteRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err = database.DeleteRecord(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Remove Record with ID: %v\n", id)
	recordsHandler(w, r)
}

func CreateProject(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var project = database.Project{
		ID:       strconv.Itoa(globalIDProject),
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
	projectsHandler(w, r)
}

func UpdateProject(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.PathValue("id")
	project := database.Project{
		ID:       strconv.Itoa(globalIDProject),
		Name:     r.FormValue("projectName"),
		Activity: r.FormValue("activity"),
		Details:  r.FormValue("details"),
	}
	err = database.UpdateProject(project)
	fmt.Printf("Update Project with ID %s\n", id)
	projectsHandler(w, r)
}

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err = database.DeleteProject(id)
	fmt.Printf("Remove Project with ID: %v\n", id)
	projectsHandler(w, r)
}
