package files

import (
	"io/ioutil"
	"os"
	"path/filepath"

	ahoy_targets "gitlab.com/hidothealth/platform/ahoy/src/target"
)

type TextFileConfig struct {
	ahoy_targets.BaseFields `mapstructure:",squash"`
	Out                     string `mapstructure:"out"`
	Content                 string `mapstructure:"content"`
}

func (tfc TextFileConfig) GetTargets(_ *ahoy_targets.TargetConfigContext) ([]*ahoy_targets.Target, error) {
	opts := []ahoy_targets.TargetOption{
		ahoy_targets.WithOuts([]string{tfc.Out}),
		ahoy_targets.WithVisibility(tfc.Visibility),
		ahoy_targets.WithLabels(tfc.Labels),
		ahoy_targets.WithTargetScript("build", &ahoy_targets.TargetScript{
			Deps: tfc.Deps,
			Run: func(target *ahoy_targets.Target, runCtx *ahoy_targets.RuntimeContext) error {
				interpolatedContent, err := target.Interpolate(tfc.Content)
				if err != nil {
					return err
				}
				return ioutil.WriteFile(filepath.Join(target.Cwd, target.Outs[0]), []byte(interpolatedContent), os.ModePerm)
			},
		}),
	}

	return []*ahoy_targets.Target{
		ahoy_targets.NewTarget(
			tfc.Name,
			opts...,
		),
	}, nil
}
