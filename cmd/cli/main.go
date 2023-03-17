package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/chelnak/ysmrr"
	"github.com/thetherington/IndexCreator/internal/helpers"
)

const FILE string = "log-syslog-warning-2021.08.22.tar.gz"
const MODE = 0755

type App struct {
	Filename   string
	Index      string
	FileDate   string
	IndexDates []time.Time
	Spinners   []*ysmrr.Spinner
	wg         sync.WaitGroup
}

func main() {
	var app App

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)

	startPtr := createCmd.String("start", "", "YYYY-MM-DD")
	endPtr := createCmd.String("end", "", "YYYY-MM-DD")

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

			go app.GenerateIndex(d, app.Spinners[x])
		}

		// wait for all to complete
		app.wg.Wait()

		sm.Stop()

	default:
		fmt.Println("expected 'create' or 'import' subcommands")
		os.Exit(1)
	}

}

func (app *App) GenerateIndex(dt time.Time, s *ysmrr.Spinner) {
	defer app.wg.Done()

	new_date := dt.Format("2006.01.02")

	r, err := os.Open(fmt.Sprintf("./%s", app.Filename))
	if err != nil {
		s.UpdateMessage(fmt.Sprintf("%s -- %s", new_date, err.Error()))
		s.Error()
		return
	}
	defer r.Close()

	path := filepath.Join(app.Index, new_date)

	// Make directory based on date and Untar the reference tar.gz file
	err = os.MkdirAll(path, MODE)
	if err != nil {
		s.UpdateMessage(fmt.Sprintf("%s -- %s", new_date, err.Error()))
		s.Error()
		return
	}

	err = helpers.Untar(path, r)
	if err != nil {
		s.UpdateMessage(fmt.Sprintf("%s -- %s", new_date, err.Error()))
		s.Error()
		return
	}

	// replace Index name to new Index name
	old_index_pattern := fmt.Sprintf("%s-%s", app.Index, app.FileDate)
	new_index_pattern := fmt.Sprintf("%s-%s", app.Index, new_date)

	s.UpdateMessage(fmt.Sprintf("%s -- Replacing (%s with %s)...", new_date, old_index_pattern, new_index_pattern))

	err = helpers.SedReplace(path, old_index_pattern, new_index_pattern)
	if err != nil {
		s.UpdateMessage(fmt.Sprintf("%s -- %s", new_date, err.Error()))
		s.Error()
		return
	}

	// replace @timestamp with new date value
	old_date_pattern := strings.ReplaceAll(app.FileDate, ".", "-")
	new_date_pattern := strings.ReplaceAll(new_date, ".", "-")

	s.UpdateMessage(fmt.Sprintf("%s -- Replacing (%s with %s)...", new_date, old_date_pattern, new_date_pattern))

	err = helpers.SedReplace(path, old_date_pattern, new_date_pattern)
	if err != nil {
		s.UpdateMessage(fmt.Sprintf("%s -- %s", new_date, err.Error()))
		s.Error()
		return
	}

	// Create new tar.gz file
	s.UpdateMessage(fmt.Sprintf("%s -- Taring...", new_date))

	w, err := os.Create(fmt.Sprintf("./%s/%s-%s.tar.gz", app.Index, app.Index, new_date))
	if err != nil {
		s.UpdateMessage(fmt.Sprintf("%s -- %s", new_date, err.Error()))
		s.Error()
		return
	}

	err = helpers.Tar(path, w)
	if err != nil {
		s.UpdateMessage(fmt.Sprintf("%s -- %s", new_date, err.Error()))
		s.Error()
		return
	}

	// Delete the work dir
	s.UpdateMessage(fmt.Sprintf("%s -- Cleaning...", new_date))

	err = os.RemoveAll(path)
	if err != nil {
		s.UpdateMessage(fmt.Sprintf("%s -- %s", new_date, err.Error()))
		s.Error()
		return
	}

	s.UpdateMessage(fmt.Sprintf("%s -- Complete", new_date))
	s.Complete()
}

func (app *App) InitDateRanges(start, end time.Time) error {
	y, m, d := start.Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	y, m, d = end.Date()
	end = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	for rd := helpers.DateRange(start, end); ; {
		date := rd()
		if date.IsZero() {
			break
		}
		app.IndexDates = append(app.IndexDates, date)
	}

	return nil
}

func (app *App) Cleanup() error {
	err := os.RemoveAll(app.Index)
	if err != nil {
		return err
	}

	return nil
}

func (app *App) ValidCreateArgs(startDate, endDate *string, createCmd *flag.FlagSet) bool {
	start, valid := helpers.ValidDateInput(startDate)
	if !valid {
		fmt.Println("Start value is invalid")
		return false
	}

	end, valid := helpers.ValidDateInput(endDate)
	if !valid {
		fmt.Println("End value is invalid")
		return false
	}

	app.InitDateRanges(start, end)

	if len(createCmd.Args()) < 1 {
		fmt.Println("No export archive provided")
		return false
	}

	// Check if the input file exists
	r, err := os.Open(createCmd.Args()[0])
	if err != nil {
		fmt.Println("Archive file provided does not exist")
		return false
	}
	r.Close()

	app.Filename = r.Name()

	// parse the file name to get the index name and date from the filename
	re := regexp.MustCompile(`(\.\/)*(.*)(\d{4}\.\d{2}\.\d{2})`)

	for _, match := range re.FindAllStringSubmatch(r.Name(), -1) {
		app.Index = strings.TrimSuffix(match[2], "-")
		app.FileDate = match[3]
	}

	return true
}

func (app *App) CreateSpinGroups() ysmrr.SpinnerManager {
	sm := ysmrr.NewSpinnerManager()

	for i := 0; i < len(app.IndexDates); i++ {
		s := sm.AddSpinner(fmt.Sprintf("%s -- Extracting...", app.IndexDates[i].Format("2006.01.02")))
		app.Spinners = append(app.Spinners, s)
	}

	return sm
}
