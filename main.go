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
	"os"

	"github.com/getlantern/systray"
	"github.com/pkg/browser"
	"github.com/spf13/viper"
)

//go:embed internal/assets/css/input.css
//go:embed internal/assets/favicon.ico
//go:embed internal/assets/js/htmx.min.js
//go:embed internal/assets/js/echarts.min.js
var content embed.FS

//internal/assets/js/echarts.js

func main() {
	// Init session
	viperInit()

	var err error
	if viper.GetBool("logfile") {
		logfile, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("could not open logfile: %v", err)
		}
		defer logfile.Close()
		log.SetOutput(logfile)
	}

	handlers.HandlerInit()

	database.Connect()
	defer database.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HomePage)
	RegisterRecordRoutes(mux)
	RegisterProjectRoutes(mux)
	RegisterOtherRoutes(mux)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(content))))

	err = mime.AddExtensionType(".js", "application/javascript")
	if err != nil {
		log.Printf("error add mime: %v", err)
	}

	if os.Getenv("TIMETRACKER_DEV") != "1" {
		go func() {
			log.Printf("Listening on http://localhost:%d\n", viper.GetInt("port"))
			if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", viper.GetInt("port")), mux); err != nil {
				log.Printf("error listening: %v", err)
			}
		}()
		//fyneSysTray()
		getlanternSysTray()
	} else {
		log.Printf("Running in DEBUG Mode")
		log.Printf("Listening on http://localhost:%d\n", viper.GetInt("port"))
		if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", viper.GetInt("port")), mux); err != nil {
			log.Printf("error listening: %v", err)
		}
	}
}

func RegisterOtherRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/currentdate", handlers.ChangeDate)
	mux.HandleFunc("GET /month", handlers.MonthlySummaryHandler)
	mux.HandleFunc("GET /year", handlers.YearlySummaryHandler)
	mux.HandleFunc("POST /api/monthlysummary", handlers.MonthlySummaryChangeMonth)
	mux.HandleFunc("POST /api/yearlysummary", handlers.YearlySummaryChangeYear)
	mux.HandleFunc("POST /api/clipboard", handlers.MonthlySummaryToClipboard)
	mux.HandleFunc("POST /api/quickbar/{name}", handlers.Quickbar)
}

func RegisterRecordRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /timeframes", handlers.RecordsPageHandler)
	mux.HandleFunc("GET /timeframes/{date}", handlers.RecordsPageDateChangeHandler)
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

func viperInit() {
	viper.SetDefault("port", 34115)
	viper.SetDefault("worktime_per_week", "39h0m0s")
	viper.SetDefault("offset_overtime", "0h0m0s")
	viper.SetDefault("logfile", false)
	viper.SetDefault("decimal_separator", ".")

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

func getlanternSysTray() {
	systray.Run(func() {
		iconBytes, err := content.ReadFile("internal/assets/favicon.ico")
		if err != nil {
			log.Println(err)
		}
		systray.SetIcon(iconBytes)
		systray.SetTitle("TimeTracker")
		systray.SetTooltip("TimeTracker")
		mBrowser := systray.AddMenuItem("Open Browser", "Open TimeTracker in Browser")
		mQuitOrig := systray.AddMenuItem("Quit", "Quit the whole app")
		url := fmt.Sprintf("http://localhost:%d", viper.GetInt("port"))

		go func() {
			<-mQuitOrig.ClickedCh
			fmt.Println("Requesting quit")
			systray.Quit()
			fmt.Println("Finished quitting")
		}()

		go func() {
			var err error
			for {
				<-mBrowser.ClickedCh
				err = browser.OpenURL(url)
				if err != nil {
					log.Println(err)
				}
			}
		}()

	}, nil)
}
