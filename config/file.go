package config

import (
	"io/ioutil"
	"strings"

	"github.com/fossas/fossa-cli/errutil"
	"github.com/fossas/fossa-cli/files"

	yaml "gopkg.in/yaml.v2"

	v1 "github.com/fossas/fossa-cli/config/file.v1"
	"github.com/fossas/fossa-cli/module"
)

// File defines the minimum interface for usage as a configuration file. We use
// an interface here in anticipation of new versions of configuration files,
// which will likely be implemented as different structs.
type File interface {
	APIKey() string
	Server() string

	Title() string
	Fetcher() string
	Project() string
	Branch() string
	Revision() string

	Modules() []module.Module
}

type NoFile struct{}

func (_ NoFile) APIKey() string {
	return ""
}

func (_ NoFile) Server() string {
	return ""
}

func (_ NoFile) Title() string {
	return ""
}

func (_ NoFile) Fetcher() string {
	return ""
}

func (_ NoFile) Project() string {
	return ""
}

func (_ NoFile) Branch() string {
	return ""
}

func (_ NoFile) Revision() string {
	return ""
}

func (_ NoFile) Modules() []module.Module {
	return []module.Module{}
}

// InitFile writes the current configuration to the current configuration file
// path.
func InitFile(modules []module.Module) File {
	// Construct module configs.
	var configs []v1.ModuleProperties
	for _, m := range modules {
		configs = append(configs, v1.ModuleProperties{
			Name:        m.Name,
			Type:        m.Type.String(),
			BuildTarget: m.BuildTarget,
			Path:        m.Dir,
			Options:     m.Options,
			Ignore:      m.Ignore,
		})
	}

	// Construct configuration file.
	return v1.File{
		Version: 1,
		CLI: v1.CLIProperties{
			Server:  Endpoint(),
			Fetcher: Fetcher(),
			Project: Project(),
		},
		Analyze: v1.AnalyzeProperties{
			Modules: configs,
		},
	}
}

func WriteFile(modules []module.Module) error {
	file := InitFile(modules)

	// Write file with header.
	data, err := yaml.Marshal(file)
	if err != nil {
		return err
	}
	configHeader := []byte(
		strings.Join([]string{
			"# Generated by FOSSA CLI (https://github.com/fossas/fossa-cli)",
			"# Visit https://fossa.io to learn more",
			"",
			"",
		}, "\n"))
	return ioutil.WriteFile(Filepath(), append(configHeader, data...), 0777)
}

func UpdateFile(modules []module.Module) error {
	return errutil.ErrNotImplemented
}

func ExistsFile() (bool, error) {
	return files.Exists(Filepath())
}
