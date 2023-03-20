package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chelnak/ysmrr"
)

const FILE string = "log-syslog-warning-2021.08.22.tar.gz"
const MODE = 0755

type App struct {
	Filename        string
	Index           string
	FileDate        string
	MaintenanceApp  string
	NodePath        string
	ElasticDumpPath string
	IndexDates      []time.Time
	Spinners        []*ysmrr.Spinner
	ImportFiles     []string
	wg              sync.WaitGroup
}

func main() {
	var app App

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)

	startPtr := createCmd.String("start", "", "YYYY-MM-DD")
	endPtr := createCmd.String("end", "", "YYYY-MM-DD")

	importCmd := flag.NewFlagSet("import", flag.ExitOnError)
	mntAppNamePtr := importCmd.String("app", "mnt-1", "inSITE Elastic Maintenance App Name (Default: mnt-1)")

	importCreateCmd := flag.NewFlagSet("createImport", flag.ExitOnError)
	impStartPtr := importCreateCmd.String("start", "", "YYYY-MM-DD")
	impEndPtr := importCreateCmd.String("end", "", "YYYY-MM-DD")
	impMntAppNamePtr := importCreateCmd.String("app", "mnt-1", "inSITE Elastic Maintenance App Name (Default: mnt-1)")

	if len(os.Args) < 2 {
		fmt.Println("expected 'create' or 'import' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "create":
		createCmd.Parse(os.Args[2:])

		if !app.ValidCreateArgs(startPtr, endPtr, createCmd) {
			os.Exit(1)
		}

		// Create spin group
		sm := app.CreateSpinGroups()

		sm.Start()

		// iterate through each date and generate the import in a go routine
		for x, d := range app.IndexDates {
			app.wg.Add(1)

			go app.GenerateIndex(d, app.Spinners[x], true)
		}

		// wait for all to complete
		app.wg.Wait()

		sm.Stop()

	case "import":
		importCmd.Parse(os.Args[2:])

		if !app.ValidImportArgs(mntAppNamePtr, importCmd) {
			os.Exit(1)
		}

		sm := app.CreateSpinGroupsImport()

		sm.Start()

		for x, d := range app.ImportFiles {
			app.wg.Add(1)

			go app.ImportIndex(d, app.Spinners[x])
		}

		// wait for all to complete
		app.wg.Wait()

		sm.Stop()

	case "createImport":
		importCreateCmd.Parse(os.Args[2:])

		if !app.ValidCreateArgs(impStartPtr, impEndPtr, importCreateCmd) {
			os.Exit(1)
		}

		if !app.ValidImportArgs(impMntAppNamePtr, importCreateCmd) {
			os.Exit(1)
		}

		// Create spin group
		sm := app.CreateSpinGroups()

		sm.Start()

		for x, d := range app.IndexDates {
			app.wg.Add(1)

			go app.CreateImportIndex(d, app.Spinners[x])
		}

		// wait for all to complete
		app.wg.Wait()

		sm.Stop()

	default:
		fmt.Println("expected 'create', 'import' or 'createImport' subcommands")
		os.Exit(1)
	}

}

func (app *App) Cleanup() error {
	err := os.RemoveAll(app.Index)
	if err != nil {
		return err
	}

	return nil
}

func (app *App) CreateSpinGroups() ysmrr.SpinnerManager {
	sm := ysmrr.NewSpinnerManager()

	for i := 0; i < len(app.IndexDates); i++ {
		s := sm.AddSpinner(fmt.Sprintf("%s -- Extracting...", app.IndexDates[i].Format("2006.01.02")))
		app.Spinners = append(app.Spinners, s)
	}

	return sm
}

func (app *App) CreateSpinGroupsImport() ysmrr.SpinnerManager {
	sm := ysmrr.NewSpinnerManager()

	for i := 0; i < len(app.ImportFiles); i++ {

		p := strings.Split(app.ImportFiles[i], "/")
		s := sm.AddSpinner(fmt.Sprintf("%s -- Extracting...", p[len(p)-1]))

		app.Spinners = append(app.Spinners, s)
	}

	return sm
}
