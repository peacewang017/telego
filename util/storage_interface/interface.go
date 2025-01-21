package storage_interface

import "telego/util"

type UserStorageInfoGetter interface {
	GetAllStorageByUser(username, password string) ([]util.UserOneStorageSet, error)
}
