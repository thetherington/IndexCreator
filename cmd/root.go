/*
Copyright © 2023 Tom Hetherington <thomas@hetheringtons.org>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/thetherington/IndexCreator/cmd/create"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "IndexCreator",
	Short: "Auto inSITE Index Creator and Importer tool",
	Long: `This is a automation tool to generate and import inSITE indexes for the sake of creating "demo data".

This application can perform the following:
  a) Auto generate index import files based on a date range.
  b) Import an inSITE index import file or a directory of import files.
  c) Auto generate import files from an inSITE index export and then auto import into Elasticsearch.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

	// err = doc.GenMarkdownTree(rootCmd, "docs")
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.AddCommand(create.CreateCmd)

	rootCmd.Version = "0.1"
}
