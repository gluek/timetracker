// Copyright 2024 Gerrit Lükens. All rights reserved.
package main

import (
	"embed"
	"fmt"
	"local/timetracker/internal/database"
	"local/timetracker/internal/handlers"
	"log"
	"mime"
	"net/http"
)

//go:embed internal/assets/css/input.css
//go:embed internal/assets/favicon.png
//go:embed internal/assets/js/htmx.min.js
//go:embed internal/assets/js/echarts.js
var content embed.FS

func main() {
	// Init session

	database.Connect()
	defer database.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.Root)
	mux.HandleFunc("/projects", handlers.ProjectsPage)
	RegisterRecordRoutes(mux)
	RegisterProjectRoutes(mux)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(content))))

	mime.AddExtensionType(".js", "application/javascript")

	fmt.Println("Listening on http://localhost:34115")
	if err := http.ListenAndServe("localhost:34115", mux); err != nil {
		log.Printf("error listening: %v", err)
	}
}

func RegisterRecordRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/timeframes", handlers.CreateRecord)
	mux.HandleFunc("DELETE /api/timeframes/{id}/", handlers.DeleteRecord)
	mux.HandleFunc("PUT /api/timeframes/{id}", handlers.UpdateRecord)
}
func RegisterProjectRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/projects", handlers.CreateProject)
	mux.HandleFunc("DELETE /api/projects/{id}/", handlers.DeleteProject)
	mux.HandleFunc("PUT /api/projects/{id}", handlers.UpdateProject)
}
func RegisterMockRecordRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/timeframes", handlers.MockCreateRecord)
	mux.HandleFunc("DELETE /api/timeframes/{id}/", handlers.MockDeleteRecord)
	mux.HandleFunc("PUT /api/timeframes/{id}", handlers.MockUpdateRecord)
}
func RegisterMockProjectRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/projects", handlers.MockCreateProject)
	mux.HandleFunc("DELETE /api/projects/{id}/", handlers.MockDeleteProject)
	mux.HandleFunc("PUT /api/projects/{id}", handlers.MockUpdateProject)
}
