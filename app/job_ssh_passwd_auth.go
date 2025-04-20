package app

import (
	"fmt"
	"os"
	"strings"
	"telego/util"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type SshPasswdAuthJob struct {
	Enable bool
}

type ModJobSshPasswdAuthStruct struct{}

var ModJobSshPasswdAuth ModJobSshPasswdAuthStruct

func (m ModJobSshPasswdAuthStruct) ConfigureSshPasswdAuth(enable bool) (string, error) {
	action := "Enabling"
	if !enable {
		action = "Disabling"
	}
	util.PrintStep("ConfigureSshPasswdAuth", fmt.Sprintf("%s SSH password authentication", action))

	// Only support Linux platforms
	if !util.IsLinux() {
		fmt.Println(color.RedString("This feature is only supported on Linux systems"))
		return "", fmt.Errorf("unsupported operating system")
	}

	// 1. Update SSH config to allow/disallow password authentication
	backupFile, err := configureSshdConfig(enable)
	if err != nil {
		return "", fmt.Errorf("failed to configure SSH server: %w", err)
	}

	// 2. Restart SSH service
	err = restartSshService()
	if err != nil {
		return backupFile, fmt.Errorf("failed to restart SSH service: %w", err)
	}

	return backupFile, nil
}

func configureSshdConfig(enable bool) (string, error) {
	sshdConfigPath := "/etc/ssh/sshd_config"

	// Check if file exists
	_, err := os.Stat(sshdConfigPath)
	if err != nil {
		return "", fmt.Errorf("SSH config file not found: %w", err)
	}

	// Read current config
	content, err := os.ReadFile(sshdConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to read SSH config: %w", err)
	}

	// Create backup
	backupPath := sshdConfigPath + ".bak." + util.CurrentTimeString()
	err = util.WriteFileWithContent(backupPath, string(content))
	if err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	util.PrintStep("ConfigureSshdConfig", fmt.Sprintf("Created backup at %s", backupPath))

	// Update configuration settings
	config := string(content)

	// Set value based on enable parameter
	value := "yes"
	if !enable {
		value = "no"
	}

	// Update PasswordAuthentication
	config = updateSshConfigSetting(config, "PasswordAuthentication", value)

	// // Update ChallengeResponseAuthentication if present
	// config = updateSshConfigSetting(config, "ChallengeResponseAuthentication", "yes")

	// // Update UsePAM if present
	// config = updateSshConfigSetting(config, "UsePAM", "yes")

	// Write updated config
	_, err = util.ModRunCmd.RequireRootRunCmd("bash", "-c", fmt.Sprintf("echo '%s' > %s", config, sshdConfigPath))
	if err != nil {
		return backupPath, fmt.Errorf("failed to write updated SSH config: %w", err)
	}

	return backupPath, nil
}

func updateSshConfigSetting(config, setting, value string) string {
	lines := strings.Split(config, "\n")
	updated := false

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), setting+" ") ||
			strings.HasPrefix(strings.TrimSpace(line), "#"+setting+" ") {
			lines[i] = setting + " " + value
			updated = true
			break
		}
	}

	// Add setting if not found
	if !updated {
		lines = append(lines, setting+" "+value)
	}

	return strings.Join(lines, "\n")
}

func restartSshService() error {
	util.PrintStep("RestartSshService", "Restarting SSH service")

	// Try systemctl first (systemd)
	_, err := util.ModRunCmd.NewBuilder("systemctl", "restart", "sshd").WithRoot().ShowProgress().BlockRun()
	if err == nil {
		return nil
	}

	// Try service command (older systems)
	_, err = util.ModRunCmd.NewBuilder("service", "sshd", "restart").WithRoot().ShowProgress().BlockRun()
	if err == nil {
		return nil
	}

	// Try ssh instead of sshd (some distributions)
	_, err = util.ModRunCmd.NewBuilder("service", "ssh", "restart").WithRoot().ShowProgress().BlockRun()
	return err
}

// Job entry point
func JobSshPasswdAuth(j SshPasswdAuthJob) error {
	// Only support Linux platforms
	if !util.IsLinux() {
		return fmt.Errorf("SSH password authentication is only supported on Linux systems")
	}

	_, err := ModJobSshPasswdAuth.ConfigureSshPasswdAuth(j.Enable)
	return err
}

func (m ModJobSshPasswdAuthStruct) NewCmd(job SshPasswdAuthJob) []string {
	cmd := []string{"telego", m.JobCmdName()}
	if !job.Enable {
		cmd = append(cmd, "--disable")
	}
	return cmd
}

func (m ModJobSshPasswdAuthStruct) JobCmdName() string {
	return "ssh-passwd-auth"
}

func (m ModJobSshPasswdAuthStruct) ParseJob(sshPasswdAuthCmd *cobra.Command) *cobra.Command {
	job := &SshPasswdAuthJob{
		Enable: true, // Default to enabling password authentication
	}

	sshPasswdAuthCmd.Flags().BoolVar(&job.Enable, "disable", false, "Disable SSH password authentication instead of enabling it")

	sshPasswdAuthCmd.Run = func(_ *cobra.Command, _ []string) {
		err := JobSshPasswdAuth(*job)
		if err != nil {
			fmt.Println(color.RedString("SSH password authentication configuration failed: %v", err))
			os.Exit(1)
		}
		if job.Enable {
			fmt.Println(color.GreenString("SSH password authentication enabled successfully"))
		} else {
			fmt.Println(color.GreenString("SSH password authentication disabled successfully"))
		}
	}

	return sshPasswdAuthCmd
}
