package app

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"telego/util"

	"github.com/fatih/color"
)

type DecodeBase64ToFileJob struct {
	Base64Content string // Base64 encoded content
	TargetPath    string // Path where to write the decoded content
	Mode          string // Optional: file mode in octal format (e.g. "0644")
}

type ModJobDecodeBase64ToFileStruct struct{}

var ModJobDecodeBase64ToFile ModJobDecodeBase64ToFileStruct

func (m ModJobDecodeBase64ToFileStruct) NewCmd(job DecodeBase64ToFileJob) []string {
	return []string{"telego", "cmd", "--cmd", "app/job_decodebase64_tofile",
		"--base64", job.Base64Content,
		"--path", job.TargetPath,
		"--mode", job.Mode}
}

func (m ModJobDecodeBase64ToFileStruct) DecodeAndWriteFile(job DecodeBase64ToFileJob) error {
	util.PrintStep("DecodeBase64ToFile", fmt.Sprintf("Decoding base64 content and writing to %s", job.TargetPath))

	// Decode base64 content
	content, err := base64.StdEncoding.DecodeString(job.Base64Content)
	if err != nil {
		return fmt.Errorf("failed to decode base64 content: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(job.TargetPath)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Determine file mode
	mode := os.FileMode(0644) // Default mode
	if job.Mode != "" {
		// Parse octal mode string
		var modeVal uint64
		_, err := fmt.Sscanf(job.Mode, "%o", &modeVal)
		if err == nil {
			mode = os.FileMode(modeVal)
		} else {
			fmt.Println(color.YellowString("Invalid mode format '%s', using default 0644", job.Mode))
		}
	}

	// Write file
	err = os.WriteFile(job.TargetPath, content, mode)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", job.TargetPath, err)
	}

	fmt.Println(color.GreenString("Successfully wrote decoded content to %s with mode %o", job.TargetPath, mode))
	return nil
}

// Job entry point
func JobDecodeBase64ToFile(j DecodeBase64ToFileJob) error {
	return ModJobDecodeBase64ToFile.DecodeAndWriteFile(j)
}
