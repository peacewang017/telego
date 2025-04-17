package app

import (
	"fmt"
	"os"
	"strings"
	"telego/util"

	"github.com/fatih/color"
)

type SshPasswdAuthJob struct{}

type ModJobSshPasswdAuthStruct struct{}

var ModJobSshPasswdAuth ModJobSshPasswdAuthStruct

func (m ModJobSshPasswdAuthStruct) NewCmd(job SshPasswdAuthJob) []string {
	return []string{"telego", "cmd", "--cmd", "app/job_ssh_passwd_auth"}
}

func (m ModJobSshPasswdAuthStruct) ConfigureSshPasswdAuth() (string, error) {
	util.PrintStep("ConfigureSshPasswdAuth", "Configuring SSH password authentication")

	// Only support Linux platforms
	if !util.IsLinux() {
		fmt.Println(color.RedString("This feature is only supported on Linux systems"))
		return "", fmt.Errorf("unsupported operating system")
	}

	// 1. Update SSH config to allow password authentication
	backupFile, err := configureSshdConfig()
	if err != nil {
		return "", fmt.Errorf("failed to configure SSH server: %w", err)
	}

	// 2. Restart SSH service
	err = restartSshService()
	if err != nil {
		return backupFile, fmt.Errorf("failed to restart SSH service: %w", err)
	}

	fmt.Println(color.GreenString("SSH password authentication configured successfully"))
	return backupFile, nil
}

func configureSshdConfig() (string, error) {
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
	err = os.WriteFile(backupPath, content, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	util.PrintStep("ConfigureSshdConfig", fmt.Sprintf("Created backup at %s", backupPath))

	// Update configuration settings
	config := string(content)

	// Update PasswordAuthentication
	config = updateSshConfigSetting(config, "PasswordAuthentication", "yes")

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

	_, err := ModJobSshPasswdAuth.ConfigureSshPasswdAuth()
	return err
}
