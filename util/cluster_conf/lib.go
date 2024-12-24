package clusterconf

import "telego/util"

type ClusterConfYmlModelGlobal struct {
	SshUser   string                      `yaml:"ssh_user"`
	SshPasswd string                      `yaml:"ssh_passwd"`
	Registry  *util.ContainerRegistryConf `yaml:"registry,omitempty"`
}

type ClusterConfYmlModelNode struct {
	Ip   string   `yaml:"ip"`
	Tags []string `yaml:"tags,omitempty"`
}

type ClusterConfYmlModel struct {
	Global ClusterConfYmlModelGlobal
	Nodes  map[string]ClusterConfYmlModelNode
}

type NodeInfo struct {
	Name string
	Ip   string
}

type NodeInfoExt struct {
	ArchOpt string
	Tags    []string
}
