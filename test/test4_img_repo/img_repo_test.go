package test3_main_node_config

import (
	"os"
	"telego/util"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestImgRepoSetup(t *testing.T) {
	// projectRoot := testutil.GetProjectRoot(t)

	// prepare main node docker with secret config
	// create ImgRepoConfig
	ymlmodel := util.ContainerRegistryConf{
		User:                      "testadmin",
		Password:                  "testpassword",
		UploaderStoreAddr:         "http://127.0.0.1:5000",
		UploaderStoreAdmin:        "testadmin",
		UploaderStoreAdminPw:      "testpassword",
		UploaderStoreTransferAddr: "http://127.0.0.1:8000",
		Tls:                       nil,
	}
	// marshal to yaml
	yml, err := yaml.Marshal(ymlmodel)
	if err != nil {
		t.Fatalf("marshal to yaml failed: %v", err)
	}
	// write to file
	err = os.WriteFile("/tmp/img_repo", yml, 0644)
	if err != nil {
		t.Fatalf("write to file failed: %v", err)
	}

	// rclone config to main node
	_, err = util.ModRunCmd.NewBuilder("rclone", "copy", "/tmp/img_repo", "remote:/teledeploy_secret").ShowProgress().BlockRun()
	if err != nil {
		t.Fatalf("rclone to main node failed: %v", err)
	}

	// telego start img repo
	_, err = util.ModRunCmd.NewBuilder("telego", "img-repo").ShowProgress().BlockRun()
	if err != nil {
		t.Fatalf("telego start img repo failed: %v", err)
	}

	// docker login
	_, err = util.ModRunCmd.NewBuilder("docker", "login", "127.0.0.1:5000", "-u", "testadmin", "-p", "testpassword").ShowProgress().BlockRun()
	if err != nil {
		t.Fatalf("docker login failed: %v", err)
	}

	t.Log("img repo started")
}
