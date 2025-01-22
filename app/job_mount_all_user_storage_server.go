package app

// 本模块为一个

import (
	"fmt"
	"net/http"
	"telego/util"
	"telego/util/gemini"
	"telego/util/storage_interface"
	"telego/util/yamlext"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

type ModJobMountAllUserStorageServerStruct struct{}

var ModJobMountAllUserStorageServer ModJobMountAllUserStorageServerStruct

func (m ModJobMountAllUserStorageServerStruct) JobCmdName() string {
	return "mount-all-user-storage-server"
}

func (m ModJobMountAllUserStorageServerStruct) getSftpServerByType(SecretConfTypeStorageViewYaml util.SecretConfTypeStorageViewYaml, sType string) (storeManageServer, storeAccessServer string, err error) {
	for _, storage := range SecretConfTypeStorageViewYaml.Storages {
		if storage.Type == sType {
			storeManageServer = storage.StoreManageServer
			storeAccessServer = storage.StoreAccessServer
			err = nil
			return
		}
	}
	err = fmt.Errorf("ModJobMountAllUserStorageServerStruct.getSftpServerByType: No sftp server found")
	return
}

func (m ModJobMountAllUserStorageServerStruct) doSftp(userStorageSets []util.UserOneStorageSet, username, password string) ([]util.UserMountsInfo, error) {
	SecretConfTypeStorageViewYaml := util.SecretConfTypeStorageViewYaml{}
	SecretConfTypeStorageViewYamlString, err := (util.MainNodeConfReader{}).ReadSecretConf(util.SecretConfTypeStorageViewYaml{})
	if err != nil {
		return nil, fmt.Errorf("ModJobMountAllUserStorageServerStruct.doSftp: Error ReadSecretConf: %v", err)
	}
	err = yamlext.UnmarshalAndValidate([]byte(SecretConfTypeStorageViewYamlString), &SecretConfTypeStorageViewYaml)
	if err != nil {
		return nil, fmt.Errorf("ModJobMountAllUserStorageServerStruct.doSftp: Error UnmarshalAndValidate: %v", err)
	}

	userMountsInfos := make([]util.UserMountsInfo, 0)
	for _, userStorageSet := range userStorageSets {
		mServer, aServer, err := m.getSftpServerByType(SecretConfTypeStorageViewYaml, userStorageSet.Type)
		if err != nil {
			return nil, fmt.Errorf("ModJobMountAllUserStorageServerStruct.doSftp: Error getSftpServerByType: %v", err)
		}
		userMountsInfos = append(userMountsInfos, util.UserMountsInfo{
			UserStorage_: userStorageSet,
			ManageServer: mServer,
			AccessServer: aServer,
		})
	}

	err = util.ModSftpgo.CreateUserSpace(SecretConfTypeStorageViewYaml, username, password, userMountsInfos)
	if err != nil {
		return nil, fmt.Errorf("ModJobMountAllUserStorageServerStruct.doSftp: Error CreateUserSpace: %v", err)
	}
	return userMountsInfos, nil
}

func (m ModJobMountAllUserStorageServerStruct) handleGetPath(c *gin.Context) {
	var req GetAllUserStorageLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request payload: %v", err),
		})
		return
	}

	gBaseUrl, err := (util.MainNodeConfReader{}).ReadSecretConf(util.SecretConfTypeGeminiAPIUrl{})
	if err != nil {
		fmt.Printf("ModJobGetAllUserStorageLinkServerStruct.handleGetPath: Error reading gemini url")
	}

	gServer, err := gemini.NewGeminiServer(gBaseUrl)
	if err != nil {
		fmt.Printf("ModJobGetAllUserStorageLinkServerStruct.handleGetPath: Error Initialize gemini server")
	}

	// 与 Gemini 交互
	userStorageSets, err := storage_interface.GetAllStorageByUser(gServer, req.UserName, req.PassWord)
	if err != nil {
		fmt.Printf("ModJobMountAllUserStorageServerStruct.handleGetPath: Error GetAllStorageByUser: %v", err)
	}

	// 返回集群信息，集群存储根目录列表
	userMountInfos, err := m.doSftp(userStorageSets, req.UserName, req.PassWord)
	if err != nil {
		fmt.Printf("ModJobMountAllUserStorageServerStruct.handleGetPath: Error doSftp: %v", err)
	}

	// 返回可挂载列表
	c.JSON(http.StatusOK, gin.H{
		"remote_infos": userMountInfos,
	})
}

func (m ModJobMountAllUserStorageServerStruct) listenRequest(port int) {
	r := gin.Default()
	r.POST("/mount_all_user_storage_server_url", m.handleGetPath)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

func (m ModJobMountAllUserStorageServerStruct) Run() {
	m.listenRequest(8083)
}

func (m ModJobMountAllUserStorageServerStruct) ParseJob(getAllUserStorageLinkServerCmd *cobra.Command) *cobra.Command {
	getAllUserStorageLinkServerCmd.Run = func(_ *cobra.Command, _ []string) {
		m.Run()
	}

	return getAllUserStorageLinkServerCmd
}

// func (m ModJobMountAllUserStorageServerStruct) SyncStoreManageServerVirtualPath(userStorages []util.UserOneStorageSet) error {
// 	ExpandUserStorageListWithCustom := func(userStorage []util.UserOneStorageSet) []util.UserOneStorageSet {
// 		return userStorage
// 	}

// 	collectMountInfo := func(userStorage []util.UserOneStorageSet) (CollectMountInfo, error) {
// 		configstr, err := util.MainNodeConfReader{}.ReadSecretConf(util.SecretConfTypeStorageViewYaml{})
// 		if err != nil {
// 			return CollectMountInfo{}, err
// 		}

// 		config := util.SecretConfTypeStorageViewYaml{}
// 		err = yamlext.UnmarshalAndValidate([]byte(configstr), &config)
// 		if err != nil {
// 			return CollectMountInfo{}, err
// 		}

// 		return CollectMountInfo{
// 			ManageAdmin:     config.StoreManageAdmin,
// 			ManageAdminPass: config.StoreManageAdminPass,
// 			EachStoreInfo: funk.Map(userStorage, func(userStorage util.UserStorage) util.UserMountsInfo {
// 				return util.UserMountsInfo{
// 					UserStorage_: userStorage,
// 					MountPath:    config.Storages[userStorage.RootStorage].MountPath,
// 					ManageServer: config.Storages[userStorage.RootStorage].StoreManageServer,
// 					AccessServer: config.Storages[userStorage.RootStorage].StoreAccessServer,
// 				}
// 			}),
// 		}, nil
// 	}
// 	// 0. 拿到 []UserStorage, 代表用户可以访问的多个存储集群以及子路径
// 	// gemeni-nm:
// 	//
// 	//	subpath1
// 	//	subpath2
// 	//	subpath3
// 	//
// 	// gemini-sh:
// 	//
// 	//	subpath1
// 	//	subpath2
// 	//	subpath3

// 	// 1. 读取配置扩展用户可以访问的目录集
// 	userStorages = ExpandUserStorageListWithCustom(userStorages)

// 	// 2. 根据配置确认每个存储的挂载路径，虚拟路径，管理员，管理员密码，管理server访问路径，数据访问路径
// 	collectedMountInfo, err := collectMountInfo(userStorages)
// 	if err != nil {
// 		return err
// 	}

// 	// 3. 将合并后的目录集写入到 storemanage-server 的虚拟路径中
// }
