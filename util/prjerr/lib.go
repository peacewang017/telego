package prjerr

import "k8s.io/kube-openapi/pkg/validation/errors"

func DistributeDeployMasterIsAlreadySetup() error {
	return errors.New(0, "DistributeDeploy - master is already setup")
}
