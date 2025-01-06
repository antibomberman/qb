package main

import (
	"log"
	"os"
)

const (
	envFile        = "dblayer.env"
	envFileContent = `DRIVER=
DSN=
MAX_ATTEMPTS=3
TIMEOUT=1
`
)

func GenerateEnv() {
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		file, err := os.Create(envFile)

		if err != nil {
			log.Fatal(err)
		}
		_, err = file.WriteString(envFileContent)
		if err != nil {
			log.Fatal(err)
		}
	}
}
