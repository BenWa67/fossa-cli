package v2

import (
	"github.com/fossas/fossa-cli/errutil"
	"github.com/fossas/fossa-cli/module"
)

type File struct {
	Version int `yaml:"version"`

	Endpoint    string `yaml:"server,omitempty"`
	Project     string `yaml:"project,omitempty"`
	Revision    string `yaml:"revision,omitempty"`
	Branch      string `yaml:"branch,omitempty"`
	ImportedVCS bool   `yaml:"imported-vcs,omitempty"`

	Modules []module.Module
}

func New(data []byte) (File, error) {
	return File{}, errutil.ErrNotImplemented
}
