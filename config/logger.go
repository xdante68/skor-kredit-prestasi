package config

import (
	"log"
	"os"
)

func Logger() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}