package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/thoas/go-funk"
)

type ModRunCmdStruct struct {
}

var ModRunCmd ModRunCmdStruct

type CmdBuilder struct {
	cmd          *exec.Cmd
	outputBuffer bytes.Buffer
	errWriters   []io.Writer
	outWriters   []io.Writer
	showProgress bool
}

func (b *CmdBuilder) Cmds() []string {
	return append([]string{b.cmd.Path}, b.cmd.Args...)
}

func (b *CmdBuilder) ShowProgress() *CmdBuilder {
	b.showProgress = true
	b.errWriters = append(b.errWriters, os.Stderr)
	b.outWriters = append(b.outWriters, os.Stdout)
	return b
}

func (b *CmdBuilder) SetDir(dir string) *CmdBuilder {
	b.cmd.Dir = dir
	return b
}

func (b *CmdBuilder) SetEnv(envs ...string) *CmdBuilder {
	env := os.Environ()
	env = append(env, envs...)
	b.cmd.Env = env
	return b
}

func (b *CmdBuilder) BlockRun() (string, error) {

	errWriter := io.MultiWriter(b.errWriters...)
	outWriter := io.MultiWriter(b.outWriters...)

	if b.showProgress {
		b.cmd.Stdout = os.Stdout
		b.cmd.Stderr = os.Stderr
	} else {
		b.cmd.Stdout = outWriter
		b.cmd.Stderr = errWriter
	}

	// 启动命令
	if err := b.cmd.Start(); err != nil {
		return "", fmt.Errorf("error starting command: %v", err)
	}

	// 等待命令执行完成
	if err := b.cmd.Wait(); err != nil {
		return "", fmt.Errorf("error waiting for command: %v", err)
	}

	return b.outputBuffer.String(), nil
}
func (b *CmdBuilder) WithRoot() *CmdBuilder {
	if IsWindows() || isRoot() {
		return b
	}
	if b.cmd.Path == "sudo" {
		return b
	}
	nb := ModRunCmd.NewBuilder("sudo", append([]string{b.cmd.Path}, b.cmd.Args...)...)

	nb.SetDir(b.cmd.Dir)
	nb.SetEnv(b.cmd.Env...)
	nb.errWriters = b.errWriters

	return b
}

func (m ModRunCmdStruct) ShowProgress(name string, args ...string) *CmdBuilder {
	b := m.NewBuilder(name, args...).ShowProgress()
	return b
}

// runCommand 用于执行 shell 命令并返回输出
func (m ModRunCmdStruct) NewBuilder(name string, args ...string) *CmdBuilder {
	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	return "", fmt.Errorf("命令执行错误: %v, 输出: %s", err, string(output))
	// }
	// return string(output), nil
	cmd := exec.Command(name, args...)

	b := CmdBuilder{
		cmd: cmd,
	}
	b.errWriters = []io.Writer{
		&b.outputBuffer,
	}
	b.outWriters = []io.Writer{
		&b.outputBuffer,
	}
	return &b
}

func (m ModRunCmdStruct) RequireRootRunCmd(name string, args ...string) (string, error) {
	// is root
	if IsWindows() || isRoot() {
		return ModRunCmd.NewBuilder(name, args...).BlockRun()
	}

	args = append([]string{name}, args...)
	return ModRunCmd.NewBuilder("sudo", args...).BlockRun()
}

func (m ModRunCmdStruct) CopyDirContentOrFileTo(srcDirOrFile, destDir string) error {
	_, err := ModRunCmd.NewBuilder("rclone", "copy", "-P", srcDirOrFile, destDir).BlockRun()
	return err
}

func (m ModRunCmdStruct) SplitCmdline(cmdline string) []string {
	return funk.Map(strings.Split(cmdline, " "), func(slice string) string {
		if slice == "\"\"" || slice == "''" {
			return ""
		}
		return slice
	}).([]string)
}

type Cmdline struct {
	Cmdline string
}

func (c Cmdline) toCmds() []string {
	return ModRunCmd.SplitCmdline(c.Cmdline)
}

type CmdModels struct {
}

func (m CmdModels) InstallTelegoWithPy() string {
	return fmt.Sprintf("python3 -c \"import urllib.request, os; script = urllib.request.urlopen('http://%s:8003/bin_telego/install.py').read(); exec(script.decode());\"", MainNodeIp)
}

func (m ModRunCmdStruct) CmdModels() CmdModels {
	return CmdModels{}
}
