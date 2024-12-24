package app

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
)

// pass pointer only
type Job interface {
	JobArgs() []string
}

func CmdsToCmd(cmds []string) string {
	return strings.Join(
		funk.Map(cmds, func(s string) string {
			if s == "" {
				return "\"\""
			} else {
				return s
			}
		}).([]string), " ",
	)
}

type JobModInterface interface {
	JobCmdName() string
	ParseJob(applyCmd *cobra.Command) *cobra.Command
}
