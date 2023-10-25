package main

import (
	"github.com/knelasevero/extended-job-manager-operator/pkg/version"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/knelasevero/extended-job-manager-operator/pkg/operator"
	"github.com/spf13/cobra"
)

func NewOperator() *cobra.Command {
	cmd := controllercmd.
		NewControllerCommandConfig("openshift-extended-job-manager-operator", version.Get(), operator.RunOperator).
		NewCommand()
	cmd.Use = "operator"
	cmd.Short = "Start the Cluster extended-job-manager Operator"

	return cmd
}
