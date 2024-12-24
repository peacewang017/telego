package app

// import (
// 	"fmt"
// 	"path/filepath"
// 	"telego/util"
// 	"testing"

// 	"k8s.io/client-go/util/homedir"
// )

// func TestGenerateSSHKey(t *testing.T) {
// 	homeDir := homedir.HomeDir()
// 	sshFile := filepath.Join(homeDir, ".ssh", "id_ed25519")
// 	cmds := []string{"bash", "-c", fmt.Sprintf("ssh-keygen -t ed25519 -f %s -N '' -q", sshFile)}
// 	t.Log("Running3", cmds)
// 	util.ModRunCmd.RunCommandShowProgress(cmds[0], cmds[1:]...)
// }
