package internal

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

func GenerateDefaultFiles() {
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		err := os.Mkdir("migrations", 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
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
