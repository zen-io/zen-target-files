package files

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	zen_targets "github.com/zen-io/zen-core/target"
	"github.com/zen-io/zen-core/utils"
	archiving "github.com/zen-io/zen-target-archiving"
)

type RemoteFileConfig struct {
	Name          string            `mapstructure:"name" desc:"Name for the target"`
	Description   string            `mapstructure:"desc" desc:"Target description"`
	Labels        []string          `mapstructure:"labels" desc:"Labels to apply to the targets"`
	Deps          []string          `mapstructure:"deps" desc:"Build dependencies"`
	PassEnv       []string          `mapstructure:"pass_env" desc:"List of environment variable names that will be passed from the OS environment, they are part of the target hash"`
	SecretEnv     []string          `mapstructure:"secret_env" desc:"List of environment variable names that will be passed from the OS environment, they are not used to calculate the target hash"`
	Env           map[string]string `mapstructure:"env" desc:"Key-Value map of static environment variables to be used"`
	Tools         map[string]string `mapstructure:"tools" desc:"Key-Value map of tools to include when executing this target. Values can be references"`
	Visibility    []string          `mapstructure:"visibility" desc:"List of visibility for this target"`
	Url           string            `mapstructure:"url"`
	Out           *string           `mapstructure:"out"`
	Hashes        []string          `mapstructure:"hashes"`
	Extract       bool              `mapstructure:"extract"`
	ExportedFiles []string          `mapstructure:"exported_files"`
	Username      string            `mapstructure:"username"`
	Password      string            `mapstructure:"password"`
	Headers       map[string]string `mapstructure:"headers"`
	Binary        bool              `mapstructure:"binary"`
}

func (rfc RemoteFileConfig) GetTargets(tcc *zen_targets.TargetConfigContext) ([]*zen_targets.Target, error) {
	return rfc.ExportTargets(tcc)
}

func (rfc RemoteFileConfig) ExportTargets(tcc *zen_targets.TargetConfigContext) ([]*zen_targets.Target, error) {
	if !rfc.Extract && len(rfc.ExportedFiles) > 0 {
		return nil, fmt.Errorf("exported files does not work without extract")
	}

	url, err := tcc.Interpolate(rfc.Url)
	if err != nil {
		return nil, err
	}
	rfc.Labels = append(rfc.Labels, fmt.Sprintf("url=%s", url))

	var downloadLocation string
	var extractOut *string

	var outs []string

	if rfc.Extract {
		var ext string
		if strings.HasSuffix(url, ".tar.gz") {
			ext = ".tar.gz"
		} else {
			ext = filepath.Ext(url)
		}

		if len(rfc.ExportedFiles) == 0 {
			if rfc.Out != nil {
				extractOut = rfc.Out
			} else {
				extractOut = utils.StringPtr(strings.TrimSuffix(filepath.Base(url), ext))
			}
			downloadLocation = "tmp_" + *extractOut + ext
			rfc.ExportedFiles = []string{"**/*"}
		} else {
			downloadLocation = filepath.Base(url)
		}

		outs = []string{downloadLocation}
	} else if len(rfc.ExportedFiles) > 0 {
		outs = rfc.ExportedFiles
	} else {
		if rfc.Out != nil {
			rfc.ExportedFiles = []string{fmt.Sprintf("%s/**/*", *rfc.Out)}
		}
		downloadLocation = filepath.Base(url)
		outs = []string{downloadLocation}
	}

	var stepName string
	if rfc.Extract {
		stepName = fmt.Sprintf("%s_download", rfc.Name)
	} else {
		stepName = rfc.Name
	}

	steps := []*zen_targets.Target{
		zen_targets.NewTarget(
			stepName,
			zen_targets.WithLabels(rfc.Labels),
			zen_targets.WithHashes(rfc.Hashes),
			zen_targets.WithOuts(outs),
			zen_targets.WithVisibility(rfc.Visibility),
			zen_targets.WithTargetScript("build", &zen_targets.TargetScript{
				Deps: rfc.Deps,
				Run: func(target *zen_targets.Target, runCtx *zen_targets.RuntimeContext) error {
					// var url string
					// for _, l := range target.Labels {
					// 	if strings.HasPrefix(l, "url=") {
					// 		if interpolatedUrl, err := target.Interpolate(strings.Split(l, "=")[1]); err != nil {
					// 			return fmt.Errorf("interpolating url: %w", err)
					// 		} else {
					// 			url = interpolatedUrl
					// 		}
					// 		break
					// 	}
					// }

					target.Debugln("Download url: %s", url)
					req, err := http.NewRequest("GET", url, nil)
					if err != nil {
						return err
					}
					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						return err
					}
					defer resp.Body.Close()

					var interpolatedDownload string
					if id, err := target.Interpolate(downloadLocation); err != nil {
						return fmt.Errorf("interpolating url: %w", err)
					} else {
						interpolatedDownload = id
					}

					f, err := os.OpenFile(filepath.Join(target.Cwd, interpolatedDownload), os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						return err
					}
					defer f.Close()

					bar := progressbar.NewOptions64(
						resp.ContentLength,
						progressbar.OptionSetDescription("Downloading..."),
						progressbar.OptionSetWriter(target),
						progressbar.OptionShowBytes(true),
						progressbar.OptionSetWidth(10),
						progressbar.OptionThrottle(100*time.Millisecond),
						progressbar.OptionShowCount(),
						// progressbar.OptionOnCompletion(func() {
						// 	ctx.Out.Write([]byte(fmt.Sprintf("Finished Downloading %s", rfc.Out)))
						// }),
						progressbar.OptionSpinnerType(14),
						progressbar.OptionFullWidth(),
					)

					if _, err := io.Copy(io.MultiWriter(f, bar), resp.Body); err != nil {
						return err
					}

					if rfc.Binary && !rfc.Extract {
						for _, o := range target.Outs {
							if err := os.Chmod(filepath.Join(target.Cwd, o), 0755); err != nil {
								return err
							}
						}
					}

					return nil
				},
			}),
		),
	}

	if rfc.Extract {
		uc := archiving.UnarchiveConfig{
			BaseFields: zen_targets.BaseFields{
				Name:       fmt.Sprintf(rfc.Name),
				Visibility: []string{":" + stepName},
				Labels:     rfc.Labels,
				Deps:       []string{fmt.Sprintf(":%s_download", rfc.Name)},
			},
			Src:           fmt.Sprintf(":%s_download", rfc.Name),
			ExportedFiles: rfc.ExportedFiles,
			Out:           extractOut,
			Binary:        rfc.Binary,
		}
		if unarchTarget, err := uc.GetTargets(tcc); err != nil {
			return nil, err
		} else {
			steps = append(steps, unarchTarget...)
		}
	}

	return steps, nil
}
