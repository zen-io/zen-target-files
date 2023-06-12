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
	archiving "github.com/tiagoposse/ahoy-archiving"
	ahoy_targets "gitlab.com/hidothealth/platform/ahoy/src/target"
	"gitlab.com/hidothealth/platform/ahoy/src/utils"
)

type RemoteFileConfig struct {
	ahoy_targets.BaseFields `mapstructure:",squash"`
	Url                     string            `mapstructure:"url"`
	Out                     *string           `mapstructure:"out"`
	Hashes                  []string          `mapstructure:"hashes"`
	Extract                 bool              `mapstructure:"extract"`
	ExportedFiles           []string          `mapstructure:"exported_files"`
	Username                string            `mapstructure:"username"`
	Password                string            `mapstructure:"password"`
	Headers                 map[string]string `mapstructure:"headers"`
	Binary                  bool              `mapstructure:"binary"`
}

func (rfc RemoteFileConfig) GetTargets(tcc *ahoy_targets.TargetConfigContext) ([]*ahoy_targets.Target, error) {
	return rfc.ExportTargets(tcc)
}

func (rfc RemoteFileConfig) ExportTargets(tcc *ahoy_targets.TargetConfigContext) ([]*ahoy_targets.Target, error) {
	if !rfc.Extract && len(rfc.ExportedFiles) > 0 {
		return nil, fmt.Errorf("exported files does not work without extract")
	}

	rfc.Labels = append(rfc.Labels, fmt.Sprintf("url=%s", rfc.Url))

	var downloadLocation string
	var extractOut *string

	var outs []string

	if rfc.Extract {
		var ext string
		if strings.HasSuffix(rfc.Url, ".tar.gz") {
			ext = ".tar.gz"
		} else {
			ext = filepath.Ext(rfc.Url)
		}

		if len(rfc.ExportedFiles) == 0 {
			if rfc.Out != nil {
				extractOut = rfc.Out
			} else {
				extractOut = utils.StringPtr(strings.TrimSuffix(filepath.Base(rfc.Url), ext))
			}
			downloadLocation = "tmp_" + *extractOut + ext
			rfc.ExportedFiles = []string{"**/*"}
		} else {
			downloadLocation = filepath.Base(rfc.Url)
		}

		outs = []string{downloadLocation}
	} else if len(rfc.ExportedFiles) > 0 {
		outs = rfc.ExportedFiles
	} else {
		if rfc.Out != nil {
			rfc.ExportedFiles = []string{fmt.Sprintf("%s/**/*", *rfc.Out)}
		}
		downloadLocation = filepath.Base(rfc.Url)
		outs = []string{downloadLocation}
	}

	var stepName string
	if rfc.Extract {
		stepName = fmt.Sprintf("%s_download", rfc.Name)
	} else {
		stepName = rfc.Name
	}

	steps := []*ahoy_targets.Target{
		ahoy_targets.NewTarget(
			stepName,
			ahoy_targets.WithLabels(rfc.Labels),
			ahoy_targets.WithHashes(rfc.Hashes),
			ahoy_targets.WithOuts(outs),
			ahoy_targets.WithVisibility(rfc.Visibility),
			ahoy_targets.WithTargetScript("build", &ahoy_targets.TargetScript{
				Deps: rfc.Deps,
				Run: func(target *ahoy_targets.Target, runCtx *ahoy_targets.RuntimeContext) error {
					var url string
					for _, l := range target.Labels {
						if strings.HasPrefix(l, "url=") {
							if interpolatedUrl, err := target.Interpolate(strings.Split(l, "=")[1]); err != nil {
								return fmt.Errorf("interpolating url: %w", err)
							} else {
								url = interpolatedUrl
							}
							break
						}
					}

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
			BaseFields: ahoy_targets.BaseFields{
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
