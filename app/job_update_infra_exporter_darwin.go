package app

import (
	"github.com/spf13/cobra"
)

type ModJobInfraExporterStruct struct{}

var ModJobInfraExporter ModJobInfraExporterStruct

func (ModJobInfraExporterStruct) ParseUpdateInfraExporter() *cobra.Command {
	return nil
}
