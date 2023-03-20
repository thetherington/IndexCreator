package app

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chelnak/ysmrr"
)

const MODE = 0755

type Config struct {
	Filename        string
	Index           string
	FileDate        string
	MaintenanceApp  string
	NodePath        string
	ElasticDumpPath string
	IndexDates      []time.Time
	Spinners        []*ysmrr.Spinner
	ImportFiles     []string
	Wg              sync.WaitGroup
}

func (config *Config) CreateSpinGroups() ysmrr.SpinnerManager {
	sm := ysmrr.NewSpinnerManager()

	for i := 0; i < len(config.IndexDates); i++ {
		s := sm.AddSpinner(fmt.Sprintf("%s -- Extracting...", config.IndexDates[i].Format("2006.01.02")))
		config.Spinners = append(config.Spinners, s)
	}

	return sm
}

func (config *Config) CreateSpinGroupsImport() ysmrr.SpinnerManager {
	sm := ysmrr.NewSpinnerManager()

	for i := 0; i < len(config.ImportFiles); i++ {

		p := strings.Split(config.ImportFiles[i], "/")
		s := sm.AddSpinner(fmt.Sprintf("%s -- Extracting...", p[len(p)-1]))

		config.Spinners = append(config.Spinners, s)
	}

	return sm
}
