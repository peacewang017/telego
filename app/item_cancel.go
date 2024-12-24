package app

func (i *MenuItem) EnterItemCancel(prefixNodes []*MenuItem) DispatchEnterNextRes {
	if len(prefixNodes) > 0 {
		// // 返回上一级菜单
		// m.Stack = m.Stack[:len(m.Stack)-1]
		// m.Current = m.Stack[len(m.Stack)-1]
		// m.Filter = ""
		// // m.input.Init()
		return DispatchEnterNextRes{
			PrevLevel: true,
		}
	} else {
		return DispatchEnterNextRes{
			Exit: true,
		}
	}

	// return true
}
