package app

import (
	"strings"
	"telego/util"
)

// prefix: deploy/k8s_{project}/apply, cur: {cluster_name}
func (selected *MenuItem) EnterItemtagExeApply(prefixNodes []*MenuItem) DispatchExecRes {
	if len(prefixNodes) < 3 {
		util.Logger.Warnf("invalid tag exe apply command %v", selected)
		return DispatchExecRes{}
	}
	stack := prefixNodes

	return DispatchExecRes{
		Exit: true,
		ExitWithDelegate: func() {
			// parent := m.ParentNode()
			// current := m.Current
			var find *MenuItem = nil
			// find node begin with bin_
			for i := len(stack) - 1; i >= 0; i-- {
				if strings.HasPrefix(stack[i].Name, "k8s_") {
					find = stack[i]
					break
				}
			}

			if find != nil {
				if find.Deployment == nil {
					// fmt.Println(color.RedString("deployment is nil for %s", find.Name))
					util.Logger.Warnf("deployment is nil for %s", find.Name)
				} else {
					ModJobApply.ApplyLocal(find.Name, find.Deployment, util.GetKubeContextByCluster(selected.Name))
				}
			} else {
				util.Logger.Warnf("can not find k8s_ prefix project")
			}
		},
	}

}
