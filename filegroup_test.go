package files

import (
	"testing"

	zen_mock "github.com/zen-io/zen-core/mock"
	"gotest.tools/v3/assert"
)

var MockSrcs = zen_mock.MockSrcsDef{
	Srcs: map[string][]string{
		"srcs": {"hello1", "hello2"},
		"bye":  {"bye*"},
	},
	SrcsMappings: map[string]map[string]string{
		"hello": {
			"hello1": "hello1",
			"hello2": "hello2",
		},
		"bye": {
			"bye1": "bye1",
		},
	},
	Outs: []string{"hello1", "bye1"},
	OutsMappings: map[string]string{
		"hello1": "hello1",
		"bye1":   "bye1",
	},
}

func TestFilegroupSimple(t *testing.T) {
	fc := &FilegroupConfig{
		Name:            "test",
		Srcs:            []string{"deep/nested/folder/*"},
		Flatten:         true,
		NoInterpolation: true,
	}
	targets, err := fc.GetTargets(nil)
	assert.NilError(t, err)
	target := targets[0]

	target.SetOriginalPath(t.TempDir())
	target.SetFqn("project", "path/to/pkg")
}
