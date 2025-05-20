package main

import (
	"embed"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/gluek/timetracker/internal/database"
	"github.com/gluek/timetracker/internal/handlers"
	"github.com/gluek/timetracker/internal/routes"

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
	routes.RegisterRecordRoutes(mux)
	routes.RegisterProjectRoutes(mux)
	routes.RegisterOtherRoutes(mux)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(content))))

	err = mime.AddExtensionType(".js", "application/javascript")
	if err != nil {
		log.Printf("error add mime: %v", err)
	}

	server := &http.Server{
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		Handler:           mux,
		Addr:              fmt.Sprintf("localhost:%d", viper.GetInt("port")),
	}

	log.Printf("Running in DEBUG Mode")
	log.Printf("Listening on http://localhost:%d\n", viper.GetInt("port"))
	if err := server.ListenAndServe(); err != nil {
		log.Printf("error listening: %v", err)
	}
}

func viperInit() {
	viper.SetDefault("port", 34115)
	viper.SetDefault("worktime_per_week", "39h0m0s")
	viper.SetDefault("offset_overtime", "0h0m0s")
	viper.SetDefault("logfile", false)
	viper.SetDefault("decimal_separator", ",")

	viper.SetConfigName("timetrackerconf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			viper.SafeWriteConfig()
			log.Println("Config file not found, creating...")
		} else {
			log.Println(err)
		}
	}
}
