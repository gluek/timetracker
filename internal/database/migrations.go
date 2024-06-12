// Copyright 2024 Gerrit Lükens. All rights reserved.
package database

import (
	"fmt"
	"log"

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
