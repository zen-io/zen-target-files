package files

import (
	"fmt"
	"path/filepath"

	ahoy_targets "gitlab.com/hidothealth/platform/ahoy/src/target"
	"gitlab.com/hidothealth/platform/ahoy/src/utils"
)

type FilegroupConfig struct {
	NoInterpolation          bool `mapstructure:"no_interpolation"`
	Flatten                  bool `mapstructure:"flatten"`
	ahoy_targets.BuildFields `mapstructure:",squash"`
}

func (fgc FilegroupConfig) GetTargets(_ *ahoy_targets.TargetConfigContext) ([]*ahoy_targets.Target, error) {
	opts := []ahoy_targets.TargetOption{
		ahoy_targets.WithSrcs(map[string][]string{"src": fgc.Srcs}),
		ahoy_targets.WithOuts(fgc.Srcs),
		ahoy_targets.WithVisibility(fgc.Visibility),
		ahoy_targets.WithTargetScript("build", &ahoy_targets.TargetScript{
			Deps: fgc.Deps,
			Run: func(target *ahoy_targets.Target, runCtx *ahoy_targets.RuntimeContext) error {
				if fgc.Flatten {
					for _, src := range target.Srcs["src"] {
						if target.ShouldInterpolate() {
							if err := ahoy_targets.CopyWithInterpolate(fmt.Sprintf("%s/%s", target.Cwd, src), fmt.Sprintf("%s/%s", target.Cwd, filepath.Base(src)), target, runCtx); err != nil {
								return err
							}
						} else if err := utils.CopyFile(fmt.Sprintf("%s/%s", target.Cwd, src), fmt.Sprintf("%s/%s", target.Cwd, filepath.Base(src))); err != nil {
							return err
						}
					}
				}

				return nil
			},
		}),
	}

	if fgc.Flatten {
		opts = append(opts, ahoy_targets.WithFlattenOuts())
	}

	if fgc.NoInterpolation {
		opts = append(opts, ahoy_targets.WithNoInterpolation())
	}

	return []*ahoy_targets.Target{
		ahoy_targets.NewTarget(
			fgc.Name,
			opts...,
		),
	}, nil
}
