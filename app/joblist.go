package app

var jobmods = []JobModInterface{
	ModJobInstall,
	ModJobApply,
	ModJobCmd,
	ModJobSsh,
	ModJobStartFileserver,
	ModJobDistributeDeploy,
	ModJobImgRepo,
	ModJobImgUploader,
	ModJobImgPrepare,
	ModJobCreateNewUser,
	ModJobFetchAdminKubeconfig,
	ModJobConfigExporter,
	ModJobRclone,
	ModJobApplyDist,
	ModJobInfraExporterSingle,
	ModJobMountAllUserStorage,
	ModJobGetAllUserStorageLinkServer,
	ModJobScanPort,
}
