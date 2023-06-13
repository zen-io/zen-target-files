package files

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	zen_targets "github.com/zen-io/zen-core/target"
)

type Replacement struct {
	From string `mapstructure:"from"`
	To   string `mapstructure:"to"`
}

type TransformConfig struct {
	Name         string            `mapstructure:"name" desc:"Name for the target"`
	Description  string            `mapstructure:"desc" desc:"Target description"`
	Labels       []string          `mapstructure:"labels" desc:"Labels to apply to the targets"`
	Deps         []string          `mapstructure:"deps" desc:"Build dependencies"`
	PassEnv      []string          `mapstructure:"pass_env" desc:"List of environment variable names that will be passed from the OS environment, they are part of the target hash"`
	SecretEnv    []string          `mapstructure:"secret_env" desc:"List of environment variable names that will be passed from the OS environment, they are not used to calculate the target hash"`
	Env          map[string]string `mapstructure:"env" desc:"Key-Value map of static environment variables to be used"`
	Tools        map[string]string `mapstructure:"tools" desc:"Key-Value map of tools to include when executing this target. Values can be references"`
	Visibility   []string          `mapstructure:"visibility" desc:"List of visibility for this target"`
	Src          string            `mapstructure:"src"`
	Out          string            `mapstructure:"out"`
	Replacements []Replacement     `mapstructure:"replacements"`
}

func (tc TransformConfig) GetTargets(_ *zen_targets.TargetConfigContext) ([]*zen_targets.Target, error) {
	opts := []zen_targets.TargetOption{
		zen_targets.WithSrcs(map[string][]string{"src": {tc.Src}}),
		zen_targets.WithOuts([]string{tc.Out}),
		zen_targets.WithVisibility(tc.Visibility),
		zen_targets.WithTargetScript("build", &zen_targets.TargetScript{
			Deps: tc.Deps,
			Run: func(target *zen_targets.Target, runCtx *zen_targets.RuntimeContext) error {
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

	return []*zen_targets.Target{
		zen_targets.NewTarget(
			tc.Name,
			opts...,
		),
	}, nil
}
