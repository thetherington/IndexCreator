/*
Copyright © 2023 Tom Hetherington <thomas@hetheringtons.org>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/thetherington/IndexCreator/internal/app"
)

var (
	mntApp string
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Subcommand used to import inSITE 'import files (tar.gz)",
	Long: `This subcommand is used to import a inSITE index import tar.gz or a directory containing import files
	
Example Usage:
  ./IndexCreator import log-syslog-informational-2023.03.15.tar.gz
  ./IndexCreator import log-syslog-informational-directory`,
	Run: func(cmd *cobra.Command, args []string) {
		var app app.Config

		if !app.ValidImportArgs(&mntApp, args) {
			os.Exit(1)
		}

		sm := app.CreateSpinGroupsImport()

		sm.Start()

		for x, d := range app.ImportFiles {
			app.Wg.Add(1)

			go app.ImportIndex(d, app.Spinners[x])
		}

		// wait for all to complete
		app.Wg.Wait()

		sm.Stop()
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	importCmd.Flags().StringVarP(&mntApp, "app", "a", "mnt-1", "inSITE Elasticsearch Maintenance Program")
}
