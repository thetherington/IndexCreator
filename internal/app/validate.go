package app

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/thetherington/IndexCreator/internal/helpers"
)

func (config *Config) InitDateRanges(start, end time.Time) error {
	y, m, d := start.Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	y, m, d = end.Date()
	end = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	for rd := helpers.DateRange(start, end); ; {
		date := rd()
		if date.IsZero() {
			break
		}
		config.IndexDates = append(config.IndexDates, date)
	}

	return nil
}

func (config *Config) ValidImportArgs(mntAppName *string, args []string) bool {
	if _, err := os.Stat(fmt.Sprintf("/opt/evertz/insite/parasite/applications/%s", *mntAppName)); os.IsNotExist(err) {
		fmt.Println("Elastic Maintenance program does not exist")
		return false
	}

	config.MaintenanceApp = *mntAppName
	config.NodePath = fmt.Sprintf("/opt/evertz/insite/parasite/applications/%s/dependencies/node/bin/node", *mntAppName)
	config.ElasticDumpPath = fmt.Sprintf("/opt/evertz/insite/parasite/applications/%s/dependencies/elasticdump/bin/elasticdump", *mntAppName)

	// validate something was provided to import
	if len(args) < 1 {
		fmt.Println("No import File or Directory provided")
		return false
	}

	// validate whatever was provided even exists
	file, err := os.Open(args[0])
	if err != nil {
		fmt.Println("Provided File or Directory does not exist")
		return false
	}
	defer file.Close()

	// parse the file name to get the index name and date from the filename
	re := regexp.MustCompile(`(\.\/)*(.*)(\d{4}\.\d{2}\.\d{2})`)

	// check whether what was provided was a directory or a file
	fileInfo, err := file.Stat()
	if err != nil {
		return false
	}

	if fileInfo.IsDir() {
		// is a Directory
		entries, err := os.ReadDir(fileInfo.Name())
		if err != nil {
			fmt.Println("Can't read directory")
			return false
		}

		for _, e := range entries {
			if re.MatchString(e.Name()) {
				config.ImportFiles = append(config.ImportFiles, filepath.Join(fileInfo.Name(), e.Name()))
			}
		}
	} else {
		// is a File
		if re.MatchString(fileInfo.Name()) {
			config.ImportFiles = append(config.ImportFiles, fileInfo.Name())
		}
	}

	if len(config.ImportFiles) < 1 {
		fmt.Println("No import files to use")
		return false
	}

	return true
}

func (config *Config) ValidCreateArgs(startDate, endDate *string, args []string) bool {
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

	config.InitDateRanges(start, end)

	if len(args) < 1 {
		fmt.Println("No export archive provided")
		return false
	}

	// Check if the input file exists
	r, err := os.Open(args[0])
	if err != nil {
		fmt.Println("Archive file provided does not exist")
		return false
	}
	r.Close()

	config.Filename = r.Name()

	// parse the file name to get the index name and date from the filename
	re := regexp.MustCompile(`(\.\/)*(.*)(\d{4}\.\d{2}\.\d{2})`)

	for _, match := range re.FindAllStringSubmatch(r.Name(), -1) {
		config.Index = strings.TrimSuffix(match[2], "-")
		config.FileDate = match[3]
	}

	return true
}
