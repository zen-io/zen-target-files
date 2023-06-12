package files

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	ahoy_targets "gitlab.com/hidothealth/platform/ahoy/src/target"
)

type Replacement struct {
	From string `mapstructure:"from"`
	To   string `mapstructure:"to"`
}

type TransformConfig struct {
	ahoy_targets.BaseFields `mapstructure:",squash"`
	Src                     string        `mapstructure:"src"`
	Out                     string        `mapstructure:"out"`
	Replacements            []Replacement `mapstructure:"replacements"`
}

func (tc TransformConfig) GetTargets(_ *ahoy_targets.TargetConfigContext) ([]*ahoy_targets.Target, error) {
	opts := []ahoy_targets.TargetOption{
		ahoy_targets.WithSrcs(map[string][]string{"src": {tc.Src}}),
		ahoy_targets.WithOuts([]string{tc.Out}),
		ahoy_targets.WithVisibility(tc.Visibility),
		ahoy_targets.WithTargetScript("build", &ahoy_targets.TargetScript{
			Deps: tc.Deps,
			Run: func(target *ahoy_targets.Target, runCtx *ahoy_targets.RuntimeContext) error {
				data, err := ioutil.ReadFile(target.Srcs["src"][0])
				if err != nil {
					return err
				}

				addVars := map[string]string{
					"CONTENT": string(data),
					"SRC":     target.Srcs["src"][0],
				}

				finalData := string(data)
				for _, rep := range tc.Replacements {
					if repFrom, err := target.Interpolate(rep.From, addVars); err != nil {
						return err
					} else if repTo, err := target.Interpolate(rep.To, addVars); err != nil {
						return err
					} else {
						finalData = strings.ReplaceAll(finalData, repFrom, repTo)
					}
				}
				return ioutil.WriteFile(filepath.Join(target.Cwd, target.Outs[0]), []byte(finalData), os.ModePerm)
			},
		}),
	}

	return []*ahoy_targets.Target{
		ahoy_targets.NewTarget(
			tc.Name,
			opts...,
		),
	}, nil
}
