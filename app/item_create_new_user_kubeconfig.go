package app

func (i *MenuItem) EnterItemCreateNewUser(prefixNodes []*MenuItem) DispatchExecRes {
	return DispatchExecRes{
		Exit: true,
		ExitWithDelegate: func() {
			CreateNewUser()
		},
	}
}
