package files

import (
	"path/filepath"

	zen_targets "github.com/zen-io/zen-core/target"
	"github.com/zen-io/zen-core/utils"
)

type ExportFileConfig struct {
	Name            string            `mapstructure:"name" desc:"Name for the target"`
	Description     string            `mapstructure:"desc" desc:"Target description"`
	Labels          []string          `mapstructure:"labels" desc:"Labels to apply to the targets"`
	Deps            []string          `mapstructure:"deps" desc:"Build dependencies"`
	PassEnv         []string          `mapstructure:"pass_env" desc:"List of environment variable names that will be passed from the OS environment, they are part of the target hash"`
	SecretEnv       []string          `mapstructure:"secret_env" desc:"List of environment variable names that will be passed from the OS environment, they are not used to calculate the target hash"`
	Env             map[string]string `mapstructure:"env" desc:"Key-Value map of static environment variables to be used"`
	Tools           map[string]string `mapstructure:"tools" desc:"Key-Value map of tools to include when executing this target. Values can be references"`
	Visibility      []string          `mapstructure:"visibility" desc:"List of visibility for this target"`
	Src             string            `mapstructure:"src"`
	Out             *string           `mapstructure:"out"`
	NoInterpolation bool              `mapstructure:"no_interpolation"`
	Binary          bool              `mapstructure:"binary"`
}

func (efc ExportFileConfig) GetTargets(_ *zen_targets.TargetConfigContext) ([]*zen_targets.Target, error) {
	if efc.Out == nil {
		baseFile := filepath.Base(efc.Src)
		efc.Out = &baseFile
	}

	opts := []zen_targets.TargetOption{
		zen_targets.WithSrcs(map[string][]string{"src": {efc.Src}}),
		zen_targets.WithOuts([]string{*efc.Out}),
		zen_targets.WithLabels(efc.Labels),
		zen_targets.WithVisibility(efc.Visibility),
		zen_targets.WithTargetScript("build", &zen_targets.TargetScript{
			Deps: efc.Deps,
			Run: func(target *zen_targets.Target, runCtx *zen_targets.RuntimeContext) error {
				from := target.Srcs["src"][0]
				to := filepath.Join(target.Cwd, target.Outs[0])

				if target.ShouldInterpolate() {
					return utils.CopyWithInterpolate(from, to, target.EnvVars())
				} else if from != to {
					return utils.CopyFile(from, to)
				}
				return nil
			},
		}),
	}
	if efc.Binary {
		opts = append(opts, zen_targets.WithBinary())
	}

	if efc.NoInterpolation {
		opts = append(opts, zen_targets.WithNoInterpolation())
	}

	return []*zen_targets.Target{
		zen_targets.NewTarget(
			efc.Name,
			opts...,
		),
	}, nil
}
