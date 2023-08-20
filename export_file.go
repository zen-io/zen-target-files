package files

import (
	"path/filepath"

	zen_targets "github.com/zen-io/zen-core/target"
)

type ExportFileConfig struct {
	Name                 string            `mapstructure:"name" zen:"yes" desc:"Name for the target"`
	Description          string            `mapstructure:"desc" zen:"yes" desc:"Target description"`
	Labels               []string          `mapstructure:"labels" zen:"yes" desc:"Labels to apply to the targets"`
	Deps                 []string          `mapstructure:"deps" zen:"yes" desc:"Build dependencies"`
	PassEnv              []string          `mapstructure:"pass_env" zen:"yes" desc:"List of environment variable names that will be passed from the OS environment, they are part of the target hash"`
	SecretEnv            []string          `mapstructure:"secret_env" desc:"List of environment variable names that will be passed from the OS environment, they are not used to calculate the target hash"`
	Env                  map[string]string `mapstructure:"env" zen:"yes" desc:"Key-Value map of static environment variables to be used"`
	Tools                map[string]string `mapstructure:"tools" zen:"yes" desc:"Key-Value map of tools to include when executing this target. Values can be references"`
	Visibility           []string          `mapstructure:"visibility" zen:"yes" desc:"List of visibility for this target"`
	Src                  string            `mapstructure:"src"`
	Out                  *string           `mapstructure:"out"`
	NoCacheInterpolation bool              `mapstructure:"no_interpolation" zen:"yes"`
	Binary               bool              `mapstructure:"binary"`
}

func (efc ExportFileConfig) GetTargets(_ *zen_targets.TargetConfigContext) ([]*zen_targets.TargetBuilder, error) {
	if efc.Out == nil {
		baseFile := filepath.Base(efc.Src)
		efc.Out = &baseFile
	}

	tb := zen_targets.ToTarget(efc)
	tb.Srcs = map[string][]string{"src": {efc.Src}}
	tb.Outs = []string{*efc.Out}
	tb.Scripts["build"] = &zen_targets.TargetBuilderScript{
		Deps:      efc.Deps,
		Env:       efc.Env,
		PassEnv:   efc.PassEnv,
		PassSecretEnv: efc.SecretEnv,
		Run: func(target *zen_targets.Target, runCtx *zen_targets.RuntimeContext) error {
			from := target.Srcs["src"][0]
			to := filepath.Join(target.Cwd, target.Outs[0])

			return target.Copy(from, to)
		},
	}

	if efc.Binary {
		tb.Binary = efc.Binary
	}

	return []*zen_targets.TargetBuilder{tb}, nil
}
