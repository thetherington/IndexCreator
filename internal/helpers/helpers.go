package helpers

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chelnak/ysmrr"
)

// Untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func Untar(dst string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	re := regexp.MustCompile(`\d{4}\.\d{2}\.\d{2}`)

	for {
		header, err := tr.Next()

		switch {
		// if no more file are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happen)
		case header == nil:
			continue

		}

		// find the date value in the file name and replace it with the value in the paths
		dir := re.FindStringSubmatch(dst)[0]
		newHeaderName := re.ReplaceAll([]byte(header.Name), []byte(dir))

		// the target location where the dir/file should be created
		target := filepath.Join(dst, string(newHeaderName))

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}

	}
}

// Tar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer; the purpose for accepting multiple writers is to allow
// for multiple outputs (for example a file, or md5 hash)
func Tar(src string, writers ...io.Writer) error {

	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("unable to tar files - %v", err.Error())
	}

	mw := io.MultiWriter(writers...)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// walk path
	return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

		// return on any error
		if err != nil {
			return err
		}

		// return on non-regular files (thanks to [kumo](https://medium.com/@komuw/just-like-you-did-fbdd7df829d3) for this suggested update)
		if !fi.Mode().IsRegular() {
			return nil
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()

		return nil
	})
}

func DateRange(start, end time.Time) func() time.Time {
	y, m, d := start.Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	y, m, d = end.Date()
	end = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	return func() time.Time {
		if start.After(end) {
			return time.Time{}
		}
		date := start
		start = start.AddDate(0, 0, 1)
		return date
	}
}

func SedReplace(path, original, new string) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	sed_replace := fmt.Sprintf("s/%s/%s/g", original, new)

	for _, item := range files {
		command := exec.Command("sed", "-i", sed_replace, filepath.Join(path, item.Name()))

		err := command.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func ValidDateInput(date *string) (time.Time, bool) {
	re := regexp.MustCompile(`\d{4}\-\d{2}\-\d{2}`)

	if re.MatchString(*date) {

		d := strings.Split(*date, "-")

		year, _ := strconv.Atoi(d[0])
		month, _ := strconv.Atoi(d[1])
		day, _ := strconv.Atoi(d[2])

		if month > 12 || day > 32 {
			return time.Now(), false
		}

		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), true
	}

	return time.Now(), false
}

func ElasticDumpRun(node_path string, args []string, s *ysmrr.Spinner, f string) error {

	command := exec.Command(node_path, args...)
	pipe, _ := command.StdoutPipe()

	command.Start()

	reader := bufio.NewReader(pipe)
	line, err := reader.ReadString('\n')

	for err == nil {
		line = strings.TrimSuffix(line, "\n")

		if strings.Contains(args[3], "data") {
			if strings.Contains(line, "offset") && strings.Contains(line, "|") {
				line = strings.Split(line, "|")[1]
				s.UpdateMessage(fmt.Sprintf("%s -- %s", f, line))
			}
		} else {
			if len(line) > 60 {
				line = line[len(line)-60:]
			}
			s.UpdateMessage(fmt.Sprintf("%s -- %s", f, line))
		}

		line, err = reader.ReadString('\n')
	}

	command.Wait()

	return nil
}
