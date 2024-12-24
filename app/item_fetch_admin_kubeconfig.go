package app

func (i *MenuItem) EnterItemFetchAdminKubeconfig(prefixNodes []*MenuItem) DispatchExecRes {
	return DispatchExecRes{
		Exit: true,
		ExitWithDelegate: func() {
			FetchAdminKubeconfig()
		},
	}
}
