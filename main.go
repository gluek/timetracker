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

	"github.com/spf13/viper"
)

//go:embed internal/assets/css/input.css
//go:embed internal/assets/favicon.png
//go:embed internal/assets/js/htmx.min.js
var content embed.FS

//internal/assets/js/echarts.js

func main() {
	// Init session
	viperDefaults()

	database.Connect()
	defer database.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HomePage)
	RegisterRecordRoutes(mux)
	RegisterProjectRoutes(mux)
	mux.HandleFunc("POST /api/currentdate", handlers.ChangeDate)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(content))))

	mime.AddExtensionType(".js", "application/javascript")

	fmt.Printf("Listening on http://localhost:%d", viper.GetInt("port"))
	if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", viper.GetInt("port")), mux); err != nil {
		log.Printf("error listening: %v", err)
	}
}

func RegisterRecordRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /timeframes", handlers.RecordsPageHandler)
	mux.HandleFunc("POST /api/timeframes", handlers.CreateRecord)
	mux.HandleFunc("DELETE /api/timeframes/{id}/", handlers.DeleteRecord)
	mux.HandleFunc("PUT /api/timeframes/{id}", handlers.UpdateRecord)
}
func RegisterProjectRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /projects", handlers.ProjectsHandler)
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

func viperDefaults() {
	viper.SetDefault("port", 34115)
	viper.SetDefault("hours_per_week", 39.0)

	viper.SetConfigName("timetracker")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			viper.SafeWriteConfig()
			log.Println("Config file not found, creating...")
		} else {
			log.Fatal(err)
		}
	}
}
