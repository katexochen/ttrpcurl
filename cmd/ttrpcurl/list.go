package main

import (
	"errors"

	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "",
		Long:  "",
		RunE:  runList,
	}

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	return errors.New("list is not implemented")
}
