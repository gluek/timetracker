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
		log.Println("Starting Migration to Version 1")
		statement, err := DB.Prepare("ALTER TABLE timeframes ADD COLUMN locationid INT DEFAULT 0 NOT NULL;")
		if err != nil {
			log.Fatal(err)
		}
		_, err = statement.Exec()
		if err != nil {
			log.Fatal(err)
		}
		setDBVersion(1)
		log.Println("Finished")
		fallthrough
	case 1:
		log.Println("Starting Migration to Version 2")
		DB.Close()
		os.Rename("timetrack.sqlite", "timetrack_old.sqlite")
		Connect()
		attach, err := DB.Prepare("ATTACH DATABASE 'timetrack_old.sqlite' AS 'old';")
		if err != nil {
			DB.Close()
			os.Rename("timetrack_old.sqlite", "timetrack.sqlite")
			panic(err)
		}
		defer attach.Close()
		_, err = attach.Exec()
		if err != nil {
			DB.Close()
			os.Rename("timetrack_old.sqlite", "timetrack.sqlite")
			panic(err)
		}

		copy, err := DB.Prepare("INSERT INTO timeframes SELECT * FROM old.timeframes;" +
			"UPDATE old.projects SET id = id + 1;" +
			"INSERT INTO projects SELECT * FROM old.projects WHERE id>4;" +
			"UPDATE old.projects SET id = id - 1;" +
			"UPDATE timeframes SET projectid = projectid + 1;" +
			"UPDATE timeframes SET locationid = locationid + 1;")
		if err != nil {
			DB.Close()
			os.Rename("timetrack_old.sqlite", "timetrack.sqlite")
			panic(err)
		}
		defer copy.Close()
		_, err = copy.Exec()
		if err != nil {
			DB.Close()
			os.Rename("timetrack_old.sqlite", "timetrack.sqlite")
			panic(err)
		}

		setDBVersion(2)
		log.Println("Finished")
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
