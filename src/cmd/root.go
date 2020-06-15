package cmd

import (
	"fmt"
	"github.com/simon-engledew/docker-context-hash/src/pkg"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
)

var cli struct {
	Dockerfile string // name of Dockerfile
	Debug      bool
}

var rootCmd = &cobra.Command{
	Use:   "context-hash <DIR>",
	Short: "Generate a repeatable hash of a Docker context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (runErr error) {
		log.SetFlags(0)
		if !cli.Debug {
			log.SetOutput(ioutil.Discard)
		}

		dir := args[0]

		hash, err := pkg.HashContext(dir, cli.Dockerfile)
		if err == nil {
			fmt.Println(hash)
		}
		return err
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cli.Dockerfile, "file", "f", "Dockerfile", "name of the Dockerfile")
	rootCmd.Flags().BoolVarP(&cli.Debug, "debug", "d", false, "list the files included in the context")

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
