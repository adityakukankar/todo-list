package utils

import "log"

const (
	HostName       string = "localhost:27017"
	DBName         string = "demo_todo"
	CollectionName string = "todo"
	Port           string = ":9000"
)

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
