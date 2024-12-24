package app

import (
	"context"
	"fmt"
	"telego/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// bf call this
//
//	make sure strings.HasPrefix(m.ParentNode().Name, "bin_")
func (i *MenuItem) EnterItemBinInstall(prefixNodes []*MenuItem) DispatchEnterNextRes {

	// select cluster to apply
	i.Children = []*MenuItem{}

	i.Children = append(i.Children, (&MenuItem{
		Name:     "local",
		Comment:  "安装到当前节点",
		Children: []*MenuItem{},
	}).setExeInstallTag())
	// i.setMultiSelectTag()
	// both k8s or bin need to select cluster
	// bin need to select node
	clusters := util.KubeList()
	for _, cluster := range clusters {

		subChildren := []*MenuItem{}
		clusterClient, err := util.KubeClusterClient(cluster)
		var finalErr error
		if err == nil {
			// 获取节点列表
			nodes, err := clusterClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			if err == nil {
				if len(nodes.Items) == 0 {
					util.Logger.Warnf("当前集群 %s 没有节点, res: %v", cluster, nodes)
				}
				for _, node := range nodes.Items {
					subChildren = append(subChildren,
						(&MenuItem{
							Name:     node.Name,
							Comment:  fmt.Sprintf("在 %s 上安装", node.Status.Addresses[0].Address),
							Children: []*MenuItem{},
						}).setExeInstallTag())
				}
			} else {
				finalErr = err
			}
		} else {
			finalErr = err
		}
		if finalErr != nil {
			util.Logger.Errorf("获取节点信息失败 %v", finalErr)
			i.Children = append(i.Children, &MenuItem{
				Name:     "warn",
				Comment:  fmt.Sprintf("获取节点信息失败 %v", finalErr),
				Children: []*MenuItem{},
			})
		} else {
			selectCluster := &MenuItem{
				Name:     cluster,
				Comment:  "选择该目标集群进行部署/安装",
				Children: subChildren,
			}
			selectCluster.setMultiSelectTag()
			i.Children = append(i.Children, selectCluster)
		}

	}
	if len(clusters) == 0 {
		util.Logger.Warnf("当前系统没有k8s集群配置")
		i.Children = append(i.Children, (&MenuItem{
			Name:     "warn",
			Comment:  "当前系统没有k8s集群配置，请先进行配置",
			Children: []*MenuItem{},
		}))
	}

	return DispatchEnterNextRes{
		NextLevel: len(i.Children) > 0,
	}
}
