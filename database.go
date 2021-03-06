package main

import (
	"fmt"
	"database/sql"
	"log"
	"strings"
)

type Database struct {
	Name string
}

type DatabaseConfig struct {
	Database []Database
}

func (d *Database) dropDatabase(tx *sql.Tx, execute bool) bool {
	query := "DROP DATABASE " + d.Name
	if execute {
		_, err := tx.Exec(query)
		if err != nil {
			return false
		}
	} else {
		fmt.Println(query)
	}

	return true
}

func (d *Database) addDatabase(tx *sql.Tx, execute bool) bool {
	query := "CREATE DATABASE " + d.Name
	if execute {
		_, err := tx.Exec(query)
		if err != nil {
			return false
		}
	} else {
		fmt.Println(query)
	}

	return true
}

func getDatabasesFromUsers(users []User) []Database {
	var databases []Database

	for _, user := range users {
		for _, permission := range user.Permissions {
			if permission.Database == "*" {
				continue
			} else {
				databases = append(databases, Database{permission.Database})
			}
		}
	}

	return databases
}

func getDatabasesFromDB(db *sql.DB) ([]Database, error)  {
	var databases []Database
	rows, err := db.Query(showAllDatabases)
	if err != nil {
		return databases, err
	}
	defer rows.Close()

	for rows.Next() {
		var database string
		if err := rows.Scan(&database); err != nil {
			log.Fatal(err)
		}
		databases = append(databases, Database{database})

	}
	if err := rows.Err(); err != nil {
		return databases, err
	}

	return databases, nil
}

func removeDuplicateDatabases(dbs []Database) []Database {
	result := []Database{}
	seen := map[Database]bool{}
	for _, db := range dbs {
		if _, ok := seen[db]; !ok {
			result = append(result, db)
			seen[db] = true
		}
	}

	return result
}

func getDatabasesToRemove(databasesFromConf, databasesFromDB []Database) []Database {
	var databasesToRemove []Database
	for _, dbDB := range databasesFromDB {
		var found bool
		for _, dbConf := range databasesFromConf {
			if dbConf.Name == dbDB.Name {
				found = true
				break
			} else if strings.Contains(dbConf.Name, "%") {
				if strings.HasPrefix(dbDB.Name, strings.Replace(dbConf.Name, "%", "", -1)) {
					found = true
					break
				}
			}
		}
		if !found {
			databasesToRemove = append(databasesToRemove, dbDB)
		}
	}

	return databasesToRemove
}

func getDatabasesToAdd(databasesFromConf, databasesFromDB []Database) []Database {
	var databasesToAdd []Database
	for _, dbConf := range databasesFromConf {
		var found bool
		for _, dbDB := range databasesFromDB {
			if dbConf.Name == dbDB.Name {
				found = true
				break
			} else if strings.Contains(dbConf.Name, "%") {
				if strings.HasPrefix(dbDB.Name, strings.Replace(dbConf.Name, "%", "", -1)) {
					found = true
					break
				}
			}
		}
		if !found && !strings.Contains(dbConf.Name, "%") {
			databasesToAdd = append(databasesToAdd, dbConf)
		}
	}

	return databasesToAdd
}