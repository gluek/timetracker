// Copyright 2024 Gerrit Lükens. All rights reserved.
package database

import (
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func Migrations() {
	version := getDBVersion()

	switch version {
	case 0:
		statement, err := DB.Prepare("ALTER TABLE timeframes ADD COLUMN locationid INT DEFAULT 0 NOT NULL;")
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
	case 1:
		DB.Close()
		os.Rename("timetrack.sqlite", "timetrack_old.sqlite")
		Connect()
		attach, err := DB.Prepare("ATTACH DATABASE 'timetrack_old.sqlite' AS 'old';")
		if err != nil {
			DB.Close()
			os.Rename("timetrack_old.sqlite", "timetrack.sqlite")
			panic(err)
		}
		_, err = attach.Exec()
		if err != nil {
			DB.Close()
			os.Rename("timetrack_old.sqlite", "timetrack.sqlite")
			panic(err)
		}

		copy, err := DB.Prepare("INSERT INTO timeframes SELECT * FROM old.timeframes; INSERT INTO projects SELECT * FROM old.projects WHERE id>3;")
		if err != nil {
			DB.Close()
			os.Rename("timetrack_old.sqlite", "timetrack.sqlite")
			panic(err)
		}
		_, err = copy.Exec()
		if err != nil {
			DB.Close()
			os.Rename("timetrack_old.sqlite", "timetrack.sqlite")
			panic(err)
		}

		setDBVersion(2)
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
