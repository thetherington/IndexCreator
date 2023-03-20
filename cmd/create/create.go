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
	start string
	end   string
)

// createCmd represents the create command
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Subcommand used to generate inSITE import files (tar.gz)",
	Long: `This subcommand is used to auto generate inSITE index import tar.gz files from a supplied start and end range
	
Example Usage:
  ./IndexCreator create --start 2023-03-20 --end 2023-03-25 log-syslog-informational-2023.03.15.tar.gz`,
	Run: func(cmd *cobra.Command, args []string) {
		var app app.Config

		if !app.ValidCreateArgs(&start, &end, args) {
			os.Exit(1)
		}

		// Create spin group
		sm := app.CreateSpinGroups()

		sm.Start()

		// iterate through each date and generate the import in a go routine
		for x, d := range app.IndexDates {
			app.Wg.Add(1)

			go app.GenerateIndex(d, app.Spinners[x], true)
		}

		// wait for all to complete
		app.Wg.Wait()

		sm.Stop()
	},
}

func init() {
	// rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	CreateCmd.Flags().StringVarP(&start, "start", "s", "", "Start Date Format (YYYY-MM-DD)")
	CreateCmd.Flags().StringVarP(&end, "end", "e", "", "End Date Format (YYYY-MM-DD)")

	CreateCmd.MarkFlagRequired("start")
	CreateCmd.MarkFlagRequired("end")
}
