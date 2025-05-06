package testutil

import (
	"os/exec"
	"testing"
)

type LinuxInstall struct {
	DefaultAppName string
	SpecAptName    string
	SpecDnfName    string
	SpecYumName    string
	SpecApkName    string
	SpecBrewName   string
}

func (l *LinuxInstall) Install(t *testing.T) {
	cmds := []string{
		"apt-get install -y " + l.DefaultAppName,
		"dnf install -y " + l.DefaultAppName,
		"yum install -y " + l.DefaultAppName,
		"apk add -y " + l.DefaultAppName,
		"brew install -y " + l.DefaultAppName,
	}

	if l.SpecAptName != "" {
		cmds[0] = "apt-get install -y " + l.SpecAptName
	}
	if l.SpecDnfName != "" {
		cmds[1] = "dnf install -y " + l.SpecDnfName
	}
	if l.SpecYumName != "" {
		cmds[2] = "yum install -y " + l.SpecYumName
	}
	if l.SpecApkName != "" {
		cmds[3] = "apk add -y " + l.SpecApkName
	}
	if l.SpecBrewName != "" {
		cmds[4] = "brew install -y " + l.SpecBrewName
	}

	for _, cmd := range cmds {
		if err := RunCommand(t, exec.Command("bash", "-c", cmd)); err != nil {
			t.Fatalf("安装 %s 失败: %v", l.DefaultAppName, err)
		} else {
			t.Logf("安装 %s 成功", l.DefaultAppName)
			return
		}
	}
}
