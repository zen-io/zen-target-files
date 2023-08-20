package files

import (
	"os"
	"path/filepath"
	"strings"

	zen_targets "github.com/zen-io/zen-core/target"
)

type FilegroupConfig struct {
	Name                 string            `mapstructure:"name" zen:"yes" desc:"Name for the target"`
	Description          string            `mapstructure:"desc" zen:"yes" desc:"Target description"`
	Labels               []string          `mapstructure:"labels" zen:"yes" desc:"Labels to apply to the targets"`
	Deps                 []string          `mapstructure:"deps" zen:"yes" desc:"Build dependencies"`
	PassEnv              []string          `mapstructure:"pass_env" zen:"yes" desc:"List of environment variable names that will be passed from the OS environment, they are part of the target hash"`
	PassSecretEnv        []string          `mapstructure:"secret_env" zen:"yes" desc:"List of environment variable names that will be passed from the OS environment, they are not used to calculate the target hash"`
	Env                  map[string]string `mapstructure:"env" zen:"yes" desc:"Key-Value map of static environment variables to be used"`
	Tools                map[string]string `mapstructure:"tools" zen:"yes" desc:"Key-Value map of tools to include when executing this target. Values can be references"`
	Visibility           []string          `mapstructure:"visibility" zen:"yes" desc:"List of visibility for this target"`
	NoCacheInterpolation bool              `mapstructure:"no_interpolation" zen:"yes"`
	Srcs                 []string          `mapstructure:"srcs" desc:"Sources for the build"`
	Outs                 []string          `mapstructure:"outs" desc:"Outs for the build"`
	Flatten              bool              `mapstructure:"flatten"`
	UnderPath            *string           `mapstructure:"under_path"`
}

func (fgc FilegroupConfig) GetTargets(_ *zen_targets.TargetConfigContext) ([]*zen_targets.TargetBuilder, error) {
	if fgc.UnderPath != nil {
		fgc.Outs = []string{*fgc.UnderPath}
	} else if fgc.Flatten {
		fgc.Outs = []string{"*"}
	} else {
		fgc.Outs = []string{"**"}
	}

	t := zen_targets.ToTarget(fgc)
	t.Srcs = map[string][]string{"src": fgc.Srcs}
	t.Outs = fgc.Outs
	t.Scripts["build"] = &zen_targets.TargetBuilderScript{
		Run: func(target *zen_targets.Target, runCtx *zen_targets.RuntimeContext) error {
			if fgc.UnderPath != nil {
				os.MkdirAll(*fgc.UnderPath, os.ModePerm)
			}

			for _, src := range target.Srcs["src"] {
				to := transformPath(strings.TrimPrefix(src, target.Cwd), fgc.UnderPath, target.Cwd, fgc.Flatten)

				if err := target.Copy(src, to); err != nil {
					return err
				}
			}

			return nil
		},
	}

	return []*zen_targets.TargetBuilder{t}, nil
}

func transformPath(p string, base_out_path *string, root string, flatten bool) string {
	if base_out_path != nil {
		root = filepath.Join(root, *base_out_path)
	}

	if flatten {
		return filepath.Join(root, filepath.Base(p))
	} else {
		return filepath.Join(root, p)
	}
}
