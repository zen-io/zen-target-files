package files

import (
	"io/ioutil"
	"os"
	"path/filepath"

	zen_targets "github.com/zen-io/zen-core/target"
)

type TextFileConfig struct {
	Name          string            `mapstructure:"name" zen:"yes" desc:"Name for the target"`
	Description   string            `mapstructure:"desc" zen:"yes" desc:"Target description"`
	Labels        []string          `mapstructure:"labels" zen:"yes" desc:"Labels to apply to the targets"`
	Deps          []string          `mapstructure:"deps" zen:"yes" desc:"Build dependencies"`
	PassEnv       []string          `mapstructure:"pass_env" zen:"yes" desc:"List of environment variable names that will be passed from the OS environment, they are part of the target hash"`
	PassSecretEnv []string          `mapstructure:"secret_env" zen:"yes" desc:"List of environment variable names that will be passed from the OS environment, they are not used to calculate the target hash"`
	Env           map[string]string `mapstructure:"env" zen:"yes" desc:"Key-Value map of static environment variables to be used"`
	Tools         map[string]string `mapstructure:"tools" zen:"yes" desc:"Key-Value map of tools to include when executing this target. Values can be references"`
	Visibility    []string          `mapstructure:"visibility" zen:"yes" desc:"List of visibility for this target"`
	Out           string            `mapstructure:"out"`
	Content       string            `mapstructure:"content"`
}

func (tfc TextFileConfig) GetTargets(_ *zen_targets.TargetConfigContext) ([]*zen_targets.TargetBuilder, error) {
	t := zen_targets.ToTarget(tfc)
	t.Outs = []string{tfc.Out}
	t.Scripts["build"] = &zen_targets.TargetBuilderScript{
		Run: func(target *zen_targets.Target, runCtx *zen_targets.RuntimeContext) error {
			interpolatedContent, err := target.Interpolate(tfc.Content)
			if err != nil {
				return err
			}
			return ioutil.WriteFile(filepath.Join(target.Cwd, target.Outs[0]), []byte(interpolatedContent), os.ModePerm)
		},
	}
	
	return []*zen_targets.TargetBuilder{t}, nil
}
