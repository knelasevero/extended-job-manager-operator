package main

import (
	"fmt"
	"os"

	"github.com/knelasevero/extended-job-manager-operator/pkg/operator"
	"github.com/knelasevero/extended-job-manager-operator/pkg/version"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/spf13/cobra"
)

func main() {
	command := NewExtendedJobManagerOperatorCommand()
	if err := command.Execute(); err != nil {
		_, err := fmt.Fprintf(os.Stderr, "%v\n", err)
		if err != nil {
			fmt.Printf("Unable to print err to stderr: %v", err)
		}
		os.Exit(1)
	}
}

func NewExtendedJobManagerOperatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extended-job-manager-operator",
		Short: "OpenShift cluster extended-job-manager operator",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				fmt.Printf("Unable to print help: %v", err)
			}
			os.Exit(1)
		},
	}

	cmd.AddCommand(NewOperator())
	return cmd
}

func NewOperator() *cobra.Command {
	cmd := controllercmd.
		NewControllerCommandConfig("openshift-extended-job-manager-operator", version.Get(), operator.RunOperator).
		NewCommand()
	cmd.Use = "operator"
	cmd.Short = "Start the Cluster extended-job-manager Operator"

	return cmd
}
