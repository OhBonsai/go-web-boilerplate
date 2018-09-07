package main

import (
	"github.com/spf13/cobra"
	"os"
)

type Command = cobra.Command

var RootCmd = &cobra.Command{
	Use:    "Bonsai",
	Short:  "go web boilerplate",
	Long:   "word-shaped wild iris leaves pierce the meadow sod just like the wistful soul beheld a simple man that impatiently rests on the threshold",
}

func Run(args []string) error {
	RootCmd.SetArgs(args)
	return RootCmd.Execute()
}

func init(){
	RootCmd.PersistentFlags().StringP("config", "c", "./config/config.json", "Configuration file to use.")
	RootCmd.PersistentFlags().Bool("watch", true, "When set config.json will be loaded from disk when the file is changed.")
}


func main(){
	if err := Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}