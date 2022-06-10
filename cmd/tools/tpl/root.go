package main

import (
	"github.com/spf13/cobra"
	"os"
)

func main() {
	cmd := &cobra.Command{
		Use: "tpl",
	}
	cmd.SetOut(os.Stdout)
	cmd.AddCommand(createRenderCommand())
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
