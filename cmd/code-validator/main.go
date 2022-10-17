package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	logger = logrus.New()
)

var (
	rootCmd = &cobra.Command{
		Use: os.Args[0],
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Info("You need to use subcommands, use --help for more information")
		},
	}
	matchCmd = &cobra.Command{
		Use:   "match",
		Short: "Exactly match 2 files (remove carriage return at end of file)",
		Long:  "exit code(0) file match\nexit code(1) Program running error, such as incorrect parameters\nexit code(2) file mismatch",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("parameter mismatch")
			}
			d1, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			d2, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}
			d1 = bytes.TrimRight(d1, "\n")
			d2 = bytes.TrimRight(d2, "\n")
			if bytes.Equal(d1, d2) {
				return nil
			} else {
				os.Exit(2)
			}

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(matchCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err)
	}
}
