package web

import (
	"net/http"

	"github.com/GoMudEngine/GoMud/internal/backup"
)

// POST /admin/api/v1/backup/download
// Admin-only. Saves all data, locks the MUD to create a tar.gz of the world
// data, then streams the archive to the client with the MUD unlocked.
func apiV1BackupDownload(w http.ResponseWriter, r *http.Request) {
	data, filename, err := backup.RunBackup()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "backup failed: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Write(data)
}
