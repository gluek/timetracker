package handlers

import (
	"fmt"
	"local/timetracker/internal/database"
	"log"
	"net/http"
)

func MockCreateRecord(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("error mock create record: %v", err)
	}
	year, month, day := parseDate(r.FormValue("dateofrecord"))
	var timeframe = database.Timeframe{
		ID:        globalID,
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
	tfList = append(tfList, timeframe)
	fmt.Printf("%s %s %d Len of tfList: %d\n", timeframe.Start, timeframe.End, timeframe.ProjectID, len(tfList))
	RecordsHandler(w, r)
}

func MockUpdateRecord(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("error mock update record: %v", err)
	}
	year, month, day := parseDate(r.FormValue("dateofrecord"))
	id := r.PathValue("id")
	index := findIDTimeframe(atoi(id))
	tfList[index] = database.Timeframe{
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
	fmt.Printf("Record with ID %s updated\n", id)
	RecordsHandler(w, r)
}

func MockDeleteRecord(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	index := findIDTimeframe(atoi(id))
	tfList = append(tfList[:index], tfList[index+1:]...)
	fmt.Printf("Remove Record with ID: %v\n", id)
	RecordsHandler(w, r)
}

func MockCreateProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("error mock create project: %v", err)
	}
	var project = database.Project{
		ID:       globalIDProject,
		Name:     r.FormValue("projectName"),
		Activity: r.FormValue("activity"),
		Details:  r.FormValue("details"),
	}
	globalIDProject += 1
	projectList = append(projectList, project)
	fmt.Printf("%s %s %s Len of projectList: %d\n", project.Activity, project.Details, project.Name, len(projectList)-1)
	ProjectsHandler(w, r)
}

func MockUpdateProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("error mock update project: %v", err)
	}
	id := r.PathValue("id")
	index := findIDProject(atoi(id))
	projectList[index] = database.Project{
		ID:       globalIDProject,
		Name:     r.FormValue("projectName"),
		Activity: r.FormValue("activity"),
		Details:  r.FormValue("details"),
	}
	fmt.Printf("Project with ID %s updated\n", id)
	ProjectsHandler(w, r)
}

func MockDeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	index := findIDProject(atoi(id))
	projectList = append(projectList[:index], projectList[index+1:]...)
	fmt.Printf("Remove Project with ID: %v\n", id)
	ProjectsHandler(w, r)
}
