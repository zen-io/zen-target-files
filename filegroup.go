package files

import (
	"path/filepath"
	"strings"

	zen_targets "github.com/zen-io/zen-core/target"
	"github.com/zen-io/zen-core/utils"
)

type FilegroupConfig struct {
	Name            string            `mapstructure:"name" desc:"Name for the target"`
	Description     string            `mapstructure:"desc" desc:"Target description"`
	Labels          []string          `mapstructure:"labels" desc:"Labels to apply to the targets"`
	Deps            []string          `mapstructure:"deps" desc:"Build dependencies"`
	PassEnv         []string          `mapstructure:"pass_env" desc:"List of environment variable names that will be passed from the OS environment, they are part of the target hash"`
	SecretEnv       []string          `mapstructure:"secret_env" desc:"List of environment variable names that will be passed from the OS environment, they are not used to calculate the target hash"`
	Env             map[string]string `mapstructure:"env" desc:"Key-Value map of static environment variables to be used"`
	Tools           map[string]string `mapstructure:"tools" desc:"Key-Value map of tools to include when executing this target. Values can be references"`
	Visibility      []string          `mapstructure:"visibility" desc:"List of visibility for this target"`
	Srcs            []string          `mapstructure:"srcs" desc:"Sources for the build"`
	Outs            []string          `mapstructure:"outs" desc:"Outs for the build"`
	NoInterpolation bool              `mapstructure:"no_interpolation"`
	Flatten         bool              `mapstructure:"flatten"`
}

func (fgc FilegroupConfig) GetTargets(_ *zen_targets.TargetConfigContext) ([]*zen_targets.Target, error) {
	var outs []string
	if fgc.Flatten {
		outs = []string{"*"}
	} else {
		outs = fgc.Srcs
	}
	opts := []zen_targets.TargetOption{
		zen_targets.WithSrcs(map[string][]string{"src": fgc.Srcs}),
		zen_targets.WithOuts(outs),
		zen_targets.WithVisibility(fgc.Visibility),
		zen_targets.WithTargetScript("build", &zen_targets.TargetScript{
			Deps: fgc.Deps,
			Run: func(target *zen_targets.Target, runCtx *zen_targets.RuntimeContext) error {
				if fgc.Flatten {
					target.Traceln("flattening srcs of %s", target.Qn())
					for _, src := range target.Srcs["src"] {
						to := filepath.Join(target.Cwd, filepath.Base(strings.TrimPrefix(src, target.Cwd)))

						if target.ShouldInterpolate() {
							if err := utils.CopyWithInterpolate(src, to, target.EnvVars()); err != nil {
								return err
							}
						} else if err := utils.CopyFile(src, to); err != nil {
							return err
						}
					}
				}

				return nil
			},
		}),
	}

	if fgc.Flatten {
		opts = append(opts, zen_targets.WithFlattenOuts())
	}

	if fgc.NoInterpolation {
		opts = append(opts, zen_targets.WithNoInterpolation())
	}

	return []*zen_targets.Target{
		zen_targets.NewTarget(
			fgc.Name,
			opts...,
		),
	}, nil
}
