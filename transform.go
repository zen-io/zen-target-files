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
	Name          string            `mapstructure:"name" zen:"yes" desc:"Name for the target"`
	Description   string            `mapstructure:"desc" zen:"yes" desc:"Target description"`
	Labels        []string          `mapstructure:"labels" zen:"yes" desc:"Labels to apply to the targets"`
	Deps          []string          `mapstructure:"deps" zen:"yes" desc:"Build dependencies"`
	PassEnv       []string          `mapstructure:"pass_env" zen:"yes" desc:"List of environment variable names that will be passed from the OS environment, they are part of the target hash"`
	PassSecretEnv []string          `mapstructure:"secret_env" zen:"yes" desc:"List of environment variable names that will be passed from the OS environment, they are not used to calculate the target hash"`
	Env           map[string]string `mapstructure:"env" zen:"yes" desc:"Key-Value map of static environment variables to be used"`
	Tools         map[string]string `mapstructure:"tools" zen:"yes" desc:"Key-Value map of tools to include when executing this target. Values can be references"`
	Visibility    []string          `mapstructure:"visibility" zen:"yes" desc:"List of visibility for this target"`
	Src           string            `mapstructure:"src"`
	Out           string            `mapstructure:"out"`
	Replacements  []Replacement     `mapstructure:"replacements"`
}

func (tc TransformConfig) GetTargets(_ *zen_targets.TargetConfigContext) ([]*zen_targets.TargetBuilder, error) {
	t := zen_targets.ToTarget(tc)
	t.Outs = []string{tc.Out}
	t.Srcs = map[string][]string{"src": {tc.Src}}
	
	t.Scripts["build"] = &zen_targets.TargetBuilderScript{
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
	}

	return []*zen_targets.TargetBuilder{t}, nil
}
