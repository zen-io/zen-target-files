package files

import (
	"fmt"
	environs "gitlab.com/hidothealth/platform/ahoy/src/environments"
	ahoy_targets "gitlab.com/hidothealth/platform/ahoy/src/target"
	"gotest.tools/v3/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
)

func TestExportFile(t *testing.T) {
	ef := &ExportFileConfig{
		BaseFields: ahoy_targets.BaseFields{
			Name: "test",
		},
		Src: "test_src",
	}

	targets, err := ef.GetTargets(nil)
	assert.NilError(t, err)

	targets[0].SetOriginalPath(t.TempDir())
	err = ioutil.WriteFile(filepath.Join(targets[0].Path(), "test_src"), []byte("NO INTERPOLATION"), os.ModePerm)
	assert.NilError(t, err)

	err = targets[0].Scripts["build"].Run(targets[0], ahoy_targets.NewRuntimeContext(
		&pflag.FlagSet{},
		make(map[string]*environs.Environment),
		targets[0].Path(),
		"",
		"",
	))
	assert.NilError(t, err)
}

// 	if efc.Out == nil {
// 		baseFile := filepath.Base(efc.Src)
// 		efc.Out = &baseFile
// 	}

// 	opts := []ahoy_targets.TargetOption{
// 		ahoy_targets.WithSrcs(map[string][]string{"src": {efc.Src}}),
// 		ahoy_targets.WithOuts([]string{*efc.Out}),
// 		ahoy_targets.WithLabels(efc.Labels),
// 		ahoy_targets.WithVisibility(efc.Visibility),
// 		ahoy_targets.WithTargetScript("build", &ahoy_targets.TargetScript{
// 			Deps: efc.Deps,
// 			Run: func(target *ahoy_targets.Target, runCtx *ahoy_targets.RuntimeContext) error {
// 				from := fmt.Sprintf("%s/%s", target.Cwd, target.Srcs["src"][0])
// 				to := fmt.Sprintf("%s/%s", target.Cwd, target.Outs[0])

// 				if target.ShouldInterpolate() {
// 					return ahoy_targets.CopyWithInterpolate(from, to, target, runCtx)
// 				} else if from != to {
// 					return utils.CopyFile(from, to)
// 				}
// 				return nil
// 			},
// 		}),
// 	}
// 	if efc.Binary {
// 		opts = append(opts, ahoy_targets.WithBinary())
// 	}

// 	if efc.NoInterpolation {
// 		opts = append(opts, ahoy_targets.WithNoInterpolation())
// 	}

// 	return []*ahoy_targets.Target{
// 		ahoy_targets.NewTarget(
// 			efc.Name,
// 			opts...,
// 		),
// 	}, nil
// }
