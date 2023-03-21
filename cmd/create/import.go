/*
Copyright © 2023 Tom Hetherington <thomas@hetheringtons.org>
*/
package create

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/thetherington/IndexCreator/internal/app"
)

var (
	importStart string
	importEnd   string
	mntApp      string
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Use this subcommand to auto import into the Elasticsearch Database after the new index data has been created",
	Long: `This subcommand is used to auto import inSITE index import files after the auto creation has completed
	
Example Usage:
  ./IndexCreator create import --start 2023-03-20 --end 2023-03-25 log-syslog-informational-2023.03.15.tar.gz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var app app.Config

		if !app.ValidCreateArgs(&importStart, &importEnd, args) {
			os.Exit(1)
		}

		if !app.ValidImportArgs(&mntApp, args) {
			os.Exit(1)
		}

		// Create spin group
		sm := app.CreateSpinGroups()

		sm.Start()

		for x, d := range app.IndexDates {
			app.Wg.Add(1)

			go app.CreateImportIndex(d, app.Spinners[x])
		}

		// wait for all to complete
		app.Wg.Wait()

		sm.Stop()
	},
}

func init() {
	CreateCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.
	importCmd.Flags().StringVarP(&importStart, "start", "s", "", "Start Date Format (YYYY-MM-DD)")
	importCmd.Flags().StringVarP(&importEnd, "end", "e", "", "End Date Format (YYYY-MM-DD)")
	importCmd.Flags().StringVarP(&mntApp, "app", "a", "mnt-1", "inSITE Elasticsearch Maintenance Program")

	importCmd.MarkFlagRequired("start")
	importCmd.MarkFlagRequired("end")
}
