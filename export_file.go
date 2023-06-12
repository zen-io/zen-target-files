package files

import (
	"path/filepath"

	ahoy_targets "gitlab.com/hidothealth/platform/ahoy/src/target"
	"gitlab.com/hidothealth/platform/ahoy/src/utils"
)

type ExportFileConfig struct {
	ahoy_targets.BaseFields `mapstructure:",squash"`
	Src                     string  `mapstructure:"src"`
	Out                     *string `mapstructure:"out"`
	NoInterpolation         bool    `mapstructure:"no_interpolation"`
	Binary                  bool    `mapstructure:"binary"`
}

func (efc ExportFileConfig) GetTargets(_ *ahoy_targets.TargetConfigContext) ([]*ahoy_targets.Target, error) {
	if efc.Out == nil {
		baseFile := filepath.Base(efc.Src)
		efc.Out = &baseFile
	}

	opts := []ahoy_targets.TargetOption{
		ahoy_targets.WithSrcs(map[string][]string{"src": {efc.Src}}),
		ahoy_targets.WithOuts([]string{*efc.Out}),
		ahoy_targets.WithLabels(efc.Labels),
		ahoy_targets.WithVisibility(efc.Visibility),
		ahoy_targets.WithTargetScript("build", &ahoy_targets.TargetScript{
			Deps: efc.Deps,
			Run: func(target *ahoy_targets.Target, runCtx *ahoy_targets.RuntimeContext) error {
				from := target.Srcs["src"][0]
				to := filepath.Join(target.Cwd, target.Outs[0])

				if target.ShouldInterpolate() {
					return ahoy_targets.CopyWithInterpolate(from, to, target, runCtx)
				} else if from != to {
					return utils.CopyFile(from, to)
				}
				return nil
			},
		}),
	}
	if efc.Binary {
		opts = append(opts, ahoy_targets.WithBinary())
	}

	if efc.NoInterpolation {
		opts = append(opts, ahoy_targets.WithNoInterpolation())
	}

	return []*ahoy_targets.Target{
		ahoy_targets.NewTarget(
			efc.Name,
			opts...,
		),
	}, nil
}
