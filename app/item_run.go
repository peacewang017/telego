package app

import "telego/util"

func (i *MenuItem) EnterItemRun(prefixNodes []*MenuItem) DispatchExecRes {
	var parentNode *MenuItem = nil
	if len(prefixNodes) >= 2 {
		parentNode = prefixNodes[len(prefixNodes)-1]
	} else {
		return DispatchExecRes{}
	}

	if parentNode.isPy() {

		// tea.Printf("start run %s/%s\n", m.CurrentPath(1), m.ParentNode().Name)
		util.CommitRunPy(prefixNodes[len(prefixNodes)-2].Name, parentNode.Name, parentNode.getEmbedScript(), nil)
		// return m, tea.Quit
		return DispatchExecRes{
			Exit: true,
			ExitWithDelegate: func() {
				util.RunPy()
			},
		}
	} else {
		util.Logger.Warnf("don't know target run type %s\n", parentNode.Name)
		return DispatchExecRes{}
	}
}
