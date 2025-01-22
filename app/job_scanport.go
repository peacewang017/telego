package app

import (
	"fmt"
	"telego/util"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
)

type ModJobScanPortStruct struct{}

var ModJobScanPort ModJobScanPortStruct

// 模式1 listcluster 列出集群
// 模式2 cluster 扫描集群cluster

func (m ModJobScanPortStruct) JobCmdName() string {
	return "scanport"
}

func (m ModJobScanPortStruct) ParseJob(scanportCmd *cobra.Command) *cobra.Command {
	listcluster := scanportCmd.Flags().BoolP("listcluster", "l", false, "list cluster")
	cluster := scanportCmd.Flags().StringP("cluster", "c", "", "cluster")
	user := scanportCmd.Flags().StringP("user", "u", "", "user")
	passwd := scanportCmd.Flags().StringP("passwd", "p", "", "passwd")
	port := scanportCmd.Flags().StringP("port", "P", "", "port")

	scanportCmd.MarkFlagsRequiredTogether("listcluster", "cluster")
	scanportCmd.Run = func(cmd *cobra.Command, args []string) {
		if *listcluster {
			err := m.Mode1ListCluster()
			if err != nil {
				fmt.Println(color.RedString("failed to list cluster: %v", err))
			} else {
				fmt.Println(color.GreenString("list cluster success"))
			}
		} else {
			err := m.Mode2ScanPort(*cluster, *user, *passwd, *port)
			if err != nil {
				fmt.Println(color.RedString("failed to scan port: %v", err))
			} else {
				fmt.Println(color.GreenString("scan port success"))
			}
		}
	}
	return scanportCmd
}

func (m ModJobScanPortStruct) Mode1ListCluster() error {
	// use kubeconfig
	clusters := util.KubeList()
	for idx, cluster := range clusters {
		fmt.Println(color.GreenString("%d. %s\n", idx+1, cluster))
	}
	return nil
}

func (m ModJobScanPortStruct) Mode2ScanPort(cluster, user, passwd, port string) error {
	ips, err := util.KubeNodeName2Ip(cluster)
	if err != nil {
		return err
	}

	hosts := funk.Map(ips, func(ip string) string {
		return user + "@" + ip
	}).([]string)

	results := util.StartRemoteCmds(hosts, "lsof -i :"+port, passwd)

	for idx, result := range results {
		fmt.Println(color.BlueString("%s : %s\n", hosts[idx], result))
	}

	return nil
}
