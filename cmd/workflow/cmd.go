package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "workflow-server",
	Long:  "",
}

var (
	debug      bool
	serverPort string
)

func init() {
	runCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", true, "debug mode")
	runCmd.PersistentFlags().StringVarP(&serverPort, "port", "p", "8080", "server port")
	runCmd.AddCommand(httpServerCmd())
}

func Execute() {
	if err := runCmd.Execute(); err != nil {
		fmt.Printf("cmd err: %s\n", err)
		os.Exit(1)
	}
}
