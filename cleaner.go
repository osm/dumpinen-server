package main

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// cleaner is responsible for deleting the files where the deleteAfter date
// has passed.
func (a *app) cleaner() {
	for {
		filesystemIDs, err := a.db.getFilesystemIDsToDelete()
		if err != nil {
			log.Printf("failed to fetch files to delete: %v\n", err)
			goto sleep
		}

		if len(filesystemIDs) == 0 {
			goto sleep
		}

		log.Printf("got %d files to delete\n", len(filesystemIDs))
		for _, filesystemID := range filesystemIDs {
			log.Printf("deleting %s\n", filesystemID)
			if err = a.db.deleteDumpByFilesystemID(filesystemID); err != nil {
				log.Printf("failed to delete file from database: %v\n", err)
				continue
			}
			if err = os.Remove(filepath.Join(a.dataDir, filesystemID)); err != nil {
				log.Printf("failed to delete file from filesystem: %v\n", err)
			}
		}
	sleep:
		time.Sleep(time.Minute * 1)
	}
}
