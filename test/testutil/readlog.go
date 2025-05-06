package testutil

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func getMostRecentLog(t *testing.T, logDir string, filter func(string) bool) string {
	files, err := filepath.Glob(path.Join(logDir, "*.log"))
	if err != nil {
		// ogger.Warnf("Error listing log files: %v", err)
		t.Fatalf("Error listing log files: %v", err)
	}
	mostRecentFile := ""
	mostRecentTime := time.Time{}
	for _, file := range files {
		if !filter(file) {
			continue
		}
		info, err := os.Stat(file)
		if err != nil {
			// Logger.Warnf("Error getting file info: %v", err)
			t.Fatalf("Error getting file info: %v", err)
		}
		if info.ModTime().After(mostRecentTime) {
			mostRecentTime = info.ModTime()
			mostRecentFile = file
		}
	}
	return mostRecentFile
}

func GetMostRecentRemoteLog(t *testing.T) string {
	return getMostRecentLog(t, "/teledeploy_secret/logs", func(file string) bool {
		return strings.HasPrefix(file, "remote_")
	})
}

func GetMostRecentLog(t *testing.T) string {
	return getMostRecentLog(t, "/teledeploy_secret/logs", func(file string) bool {
		return !strings.HasPrefix(file, "remote_")
	})
}
