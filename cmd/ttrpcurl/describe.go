package main

import (
	"errors"

	"github.com/spf13/cobra"
)

func newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "",
		Long:  "",
		RunE:  runDescribe,
	}

	return cmd
}

func runDescribe(cmd *cobra.Command, args []string) error {
	return errors.New("describe is not implemented")
}
