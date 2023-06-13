package files

import (
	"io/ioutil"
	"os"
	"path/filepath"

	zen_targets "github.com/zen-io/zen-core/target"
)

type TextFileConfig struct {
	Name        string            `mapstructure:"name" desc:"Name for the target"`
	Description string            `mapstructure:"desc" desc:"Target description"`
	Labels      []string          `mapstructure:"labels" desc:"Labels to apply to the targets"`
	Deps        []string          `mapstructure:"deps" desc:"Build dependencies"`
	PassEnv     []string          `mapstructure:"pass_env" desc:"List of environment variable names that will be passed from the OS environment, they are part of the target hash"`
	SecretEnv   []string          `mapstructure:"secret_env" desc:"List of environment variable names that will be passed from the OS environment, they are not used to calculate the target hash"`
	Env         map[string]string `mapstructure:"env" desc:"Key-Value map of static environment variables to be used"`
	Tools       map[string]string `mapstructure:"tools" desc:"Key-Value map of tools to include when executing this target. Values can be references"`
	Visibility  []string          `mapstructure:"visibility" desc:"List of visibility for this target"`
	Out         string            `mapstructure:"out"`
	Content     string            `mapstructure:"content"`
}

func (tfc TextFileConfig) GetTargets(_ *zen_targets.TargetConfigContext) ([]*zen_targets.Target, error) {
	opts := []zen_targets.TargetOption{
		zen_targets.WithOuts([]string{tfc.Out}),
		zen_targets.WithVisibility(tfc.Visibility),
		zen_targets.WithLabels(tfc.Labels),
		zen_targets.WithTargetScript("build", &zen_targets.TargetScript{
			Deps: tfc.Deps,
			Run: func(target *zen_targets.Target, runCtx *zen_targets.RuntimeContext) error {
				interpolatedContent, err := target.Interpolate(tfc.Content)
				if err != nil {
					return err
				}
				return ioutil.WriteFile(filepath.Join(target.Cwd, target.Outs[0]), []byte(interpolatedContent), os.ModePerm)
			},
		}),
	}

	return []*zen_targets.Target{
		zen_targets.NewTarget(
			tfc.Name,
			opts...,
		),
	}, nil
}
