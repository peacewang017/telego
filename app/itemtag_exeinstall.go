package app

import (
	"fmt"
	"strings"
	"telego/util"

	"github.com/fatih/color"
)

// prefix: {deploy || deploy-template}/bin_{project}/install, cur: {local || cluster_name/some_node}
func (selected *MenuItem) EnterItemtagExeInstall(prefixNodes []*MenuItem) DispatchExecRes {
	var parentNode *MenuItem = nil
	if len(prefixNodes) > 2 {
		parentNode = prefixNodes[len(prefixNodes)-1]
	} else {
		util.Logger.Warnf("invalid tag exe install command %v", selected)
		return DispatchExecRes{}
	}

	return DispatchExecRes{
		Exit: true,
		ExitWithDelegate: func() {
			parent := parentNode
			current := parentNode
			stack := prefixNodes[:len(prefixNodes)-1]

			var find *MenuItem = nil
			// find node begin with bin_
			for i := len(stack) - 1; i >= 0; i-- {
				if strings.HasPrefix(stack[i].Name, "bin_") {
					find = stack[i]
					break
				}
			}

			if find != nil {
				if selected.Name == "local" {
					// pos = append(pos, "local")
					ModJobInstall.InstallLocal(find.Name)
				} else {
					nodes := []string{}
					for _, node := range current.Children {
						if node.isSelected() {
							nodes = append(nodes, strings.ReplaceAll(node.Name, " ✓", ""))
						}
					}
					ModJobInstall.InstallToNodes(find.Name, parent.Name, nodes)
				}
			} else {
				fmt.Println(color.RedString("logical bug, install not under bin_ project"))
			}

			// if find != nil {
			// 	fmt.Printf("install %s to %v\n", find.Name, pos)
			// }
		},
	}

}
