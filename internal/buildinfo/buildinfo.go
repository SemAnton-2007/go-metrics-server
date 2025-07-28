package buildinfo

import "log"

var (
	Version = "N/A"
	Date    = "N/A"
	Commit  = "N/A"
)

func Print() {
	log.Printf("Build version: %s", Version)
	log.Printf("Build date: %s", Date)
	log.Printf("Build commit: %s", Commit)
}
