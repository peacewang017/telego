package util

import (
	"testing"
	"time"
)

// Start integration test sftpgo server
//
// docker run -d \
//   --name sftpgo \
//   -p 2022:2022 \
//   -p 8080:8080 \
//   -e SFTPGO_DEFAULT_ADMIN_USERNAME=admin \
//   -e SFTPGO_DEFAULT_ADMIN_PASSWORD=yourpassword \
//   -v ./app:/app \
//   -v ./util:/util \
//   drakkan/sftpgo:v2.6.2

// Test with cmd
// go test -v -run TestStoreserverSftpgo telego/util

func TestStoreserverSftpgo(t *testing.T) {
	err := ModSftpgo.CreateUserSpace(SecretConfTypeStorageViewYaml{
		// "admin", "yourpassword"
		StoreManageAdmin:     "admin",
		StoreManageAdminPass: "yourpassword",
		Storages: map[string]StorageViewYamlModelOneStore{
			"teststore": {
				Type:              "gemini",
				StoreManageServer: "http://localhost:8080",
				StoreAccessServer: "127.0.0.1:2022",
				MountPath:         "/app",
			},
			"teststore2": {
				Type:              "gemini",
				StoreManageServer: "http://localhost:8080",
				StoreAccessServer: "127.0.0.1:2022",
				MountPath:         "/util",
			},
		},
	}, "pa1", "991213", []UserMountsInfo{
		{
			AccessServer: "127.0.0.1:2022",
			ManageServer: "http://localhost:8080",
			UserStorage_: UserOneStorageSet{
				RootStorage: "/app",
				SubPaths: []string{
					"config",
					"notexist",
					// "dist",
					// "startup_steps",
				},
			},
		},
		{
			AccessServer: "127.0.0.1:2022",
			ManageServer: "http://localhost:8080",
			UserStorage_: UserOneStorageSet{
				RootStorage: "/util",
				SubPaths: []string{
					"gemini",
					"strext",
					// "dist",
					// "startup_steps",
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// sleep for long time
	time.Sleep(1000 * time.Hour)
}
