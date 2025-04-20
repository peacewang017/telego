package app

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"telego/util"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type DecodeBase64ToFileJob struct {
	Base64Content string // Base64 encoded content
	TargetPath    string // Path where to write the decoded content
	Mode          string // Optional: file mode in octal format (e.g. "0644")
}

type ModJobDecodeBase64ToFileStruct struct{}

var ModJobDecodeBase64ToFile ModJobDecodeBase64ToFileStruct

func (m ModJobDecodeBase64ToFileStruct) NewCmd(job DecodeBase64ToFileJob) []string {
	return []string{"telego", m.JobCmdName(),
		"--base64", job.Base64Content,
		"--path", job.TargetPath,
		"--mode", job.Mode}
}

func (m ModJobDecodeBase64ToFileStruct) JobCmdName() string {
	return "decode-base64-to-file"
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

func (m ModJobDecodeBase64ToFileStruct) ParseJob(decodeBase64Cmd *cobra.Command) *cobra.Command {
	job := &DecodeBase64ToFileJob{
		Mode: "0644", // Default mode
	}

	decodeBase64Cmd.Flags().StringVar(&job.Base64Content, "base64", "", "Base64 encoded content to write")
	decodeBase64Cmd.Flags().StringVar(&job.TargetPath, "path", "", "Path where to write the decoded content")
	decodeBase64Cmd.Flags().StringVar(&job.Mode, "mode", "0644", "File mode in octal format (e.g. 0644)")

	// Mark required flags
	_ = decodeBase64Cmd.MarkFlagRequired("base64")
	_ = decodeBase64Cmd.MarkFlagRequired("path")

	decodeBase64Cmd.Run = func(_ *cobra.Command, _ []string) {
		err := JobDecodeBase64ToFile(*job)
		if err != nil {
			fmt.Println(color.RedString("Failed to decode and write file: %v", err))
			os.Exit(1)
		}
		fmt.Println(color.GreenString("Successfully wrote file to %s", job.TargetPath))
	}

	return decodeBase64Cmd
}
