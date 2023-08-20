package files

// import (
// 	"io/ioutil"
// 	"os"
// 	"path/filepath"
// 	"testing"

// 	environs "github.com/zen-io/zen-core/environments"
// 	zen_targets "github.com/zen-io/zen-core/target"
// 	"gotest.tools/v3/assert"

// 	"github.com/spf13/pflag"
// )

// func TestExportFile(t *testing.T) {
// 	ef := &ExportFileConfig{
// 		Name: "test",
// 		Src:  "test_src",
// 	}

// 	targets, err := ef.GetTargets(nil)
// 	assert.NilError(t, err)

// 	targets[0].SetOriginalPath(t.TempDir())
// 	err = ioutil.WriteFile(filepath.Join(targets[0].Path(), "test_src"), []byte("NO INTERPOLATION"), os.ModePerm)
// 	assert.NilError(t, err)

// 	err = targets[0].Scripts["build"].Run(targets[0], zen_targets.NewRuntimeContext(
// 		&pflag.FlagSet{},
// 		make(map[string]*environs.Environment),
// 		"",
// 		"",
// 		"",
// 	))

// 	assert.NilError(t, err)
// }

// // 	if efc.Out == nil {
// // 		baseFile := filepath.Base(efc.Src)
// // 		efc.Out = &baseFile
// // 	}

// // 	opts := []zen_targets.TargetOption{
// // 		zen_targets.WithSrcs(map[string][]string{"src": {efc.Src}}),
// // 		zen_targets.WithOuts([]string{*efc.Out}),
// // 		zen_targets.WithLabels(efc.Labels),
// // 		zen_targets.WithVisibility(efc.Visibility),
// // 		zen_targets.WithTargetScript("build", &zen_targets.TargetScript{
// // 			Deps: efc.Deps,
// // 			Run: func(target *zen_targets.Target, runCtx *zen_targets.RuntimeContext) error {
// // 				from := fmt.Sprintf("%s/%s", target.Cwd, target.Srcs["src"][0])
// // 				to := fmt.Sprintf("%s/%s", target.Cwd, target.Outs[0])

// // 				if target.ShouldInterpolate() {
// // 					return zen_targets.CopyWithInterpolate(from, to, target, runCtx)
// // 				} else if from != to {
// // 					return utils.CopyFile(from, to)
// // 				}
// // 				return nil
// // 			},
// // 		}),
// // 	}
// // 	if efc.Binary {
// // 		opts = append(opts, zen_targets.WithBinary())
// // 	}

// // 	if efc.NoInterpolation {
// // 		opts = append(opts, zen_targets.WithNoInterpolation())
// // 	}

// // 	return []*zen_targets.Target{
// // 		zen_targets.NewTarget(
// // 			efc.Name,
// // 			opts...,
// // 		),
// // 	}, nil
// // }
