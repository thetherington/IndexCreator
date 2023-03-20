package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chelnak/ysmrr"
	"github.com/thetherington/IndexCreator/internal/helpers"
)

func (app *App) CreateImportIndex(dt time.Time, s *ysmrr.Spinner) {
	defer app.wg.Done()

	app.wg.Add(1)

	path, date := app.GenerateIndex(dt, s)

	for _, arg := range []string{"settings", "mapping", "data"} {

		s.UpdateMessage(fmt.Sprintf("%s -- %s", date, fmt.Sprintf("Importing %s...", arg)))

		args := []string{
			app.ElasticDumpPath,
			fmt.Sprintf("--input=%s", filepath.Join(path, fmt.Sprintf("%s-%s-%s.json", app.Index, date, arg))),
			fmt.Sprintf("--output=http://localhost:9200/%s-%s", app.Index, date),
			fmt.Sprintf("--type=%s", arg),
			"--concurrencyInterval=500",
			"--limit=1000",
			"--intervalCap=10",
		}

		// fmt.Println(args)
		helpers.ElasticDumpRun(app.NodePath, args, s, date)

	}

	// Delete the work dir
	s.UpdateMessage(fmt.Sprintf("%s -- Cleaning...", date))

	err := os.RemoveAll(path)
	if err != nil {
		s.UpdateMessage(fmt.Sprintf("%s -- %s", date, err.Error()))
		s.Error()
		return
	}

	s.UpdateMessage(fmt.Sprintf("%s -- Complete", date))
	s.Complete()
}

func (app *App) ImportIndex(f string, s *ysmrr.Spinner) {
	defer app.wg.Done()

	r, err := os.Open(fmt.Sprintf("./%s", f))
	if err != nil {
		s.UpdateMessage(fmt.Sprintf("%s -- %s", f, err.Error()))
		s.Error()
		return
	}
	defer r.Close()

	path := strings.TrimSuffix(f, ".tar.gz")

	// Make directory based on date and Untar the reference tar.gz file
	err = os.MkdirAll(path, MODE)
	if err != nil {
		s.UpdateMessage(fmt.Sprintf("%s -- %s", f, err.Error()))
		s.Error()
		return
	}

	err = helpers.Untar(path, r)
	if err != nil {
		s.UpdateMessage(fmt.Sprintf("%s -- %s", f, err.Error()))
		s.Error()
		return
	}

	// scan through each file (in order) and run ElasticDump
	path_slice := strings.Split(path, "/")
	index := path_slice[len(path_slice)-1]

	for _, arg := range []string{"settings", "mapping", "data"} {

		s.UpdateMessage(fmt.Sprintf("%s -- %s", f, fmt.Sprintf("Importing %s...", arg)))

		args := []string{
			app.ElasticDumpPath,
			fmt.Sprintf("--input=%s", filepath.Join(path, fmt.Sprintf("%s-%s.json", index, arg))),
			fmt.Sprintf("--output=http://localhost:9200/%s", index),
			fmt.Sprintf("--type=%s", arg),
			"--concurrencyInterval=500",
			"--limit=1000",
			"--intervalCap=10",
		}

		helpers.ElasticDumpRun(app.NodePath, args, s, f)

	}

	// Delete the work dir
	s.UpdateMessage(fmt.Sprintf("%s -- Cleaning...", f))

	err = os.RemoveAll(path)
	if err != nil {
		s.UpdateMessage(fmt.Sprintf("%s -- %s", f, err.Error()))
		s.Error()
		return
	}

	s.UpdateMessage(fmt.Sprintf("%s -- Complete", f))
	s.Complete()
}

func (app *App) GenerateIndex(dt time.Time, s *ysmrr.Spinner, cleanup ...bool) (path string, new_date string) {
	defer app.wg.Done()

	new_date = dt.Format("2006.01.02")

	r, err := os.Open(fmt.Sprintf("./%s", app.Filename))
	if err != nil {
		s.UpdateMessage(fmt.Sprintf("%s -- %s", new_date, err.Error()))
		s.Error()
		return
	}
	defer r.Close()

	path = filepath.Join(app.Index, new_date)

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

	if len(cleanup) > 0 {

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

	return
}
