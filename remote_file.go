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
	urlExt := utils.FileExtension(url)

	rfc.Labels = append(rfc.Labels, fmt.Sprintf("url=%s", url))

	var mainStepName, downloadLocation string

	steps := make([]*zen_targets.Target, 0)
	if rfc.Extract {
		mainStepName = fmt.Sprintf("%s_download", rfc.Name)
		mainStepRef := fmt.Sprintf(":%s", mainStepName)
		var outs []string
		if len(rfc.ExportedFiles) > 0 {
			outs = rfc.ExportedFiles
		} else {
			outs = []string{"**/*"}
		}

		if unarchTarget, err := (archiving.UnarchiveConfig{
			Name:          rfc.Name,
			Visibility:    []string{mainStepRef},
			Labels:        rfc.Labels,
			Deps:          []string{mainStepRef},
			Src:           mainStepRef,
			ExportedFiles: outs,
			Binary:        rfc.Binary,
		}).GetTargets(tcc); err != nil {
			return nil, err
		} else {
			steps = append(steps, unarchTarget...)
		}

		downloadLocation = "download" + urlExt
	} else {
		mainStepName = rfc.Name

		if rfc.Out != nil {
			downloadLocation = *rfc.Out
		} else {
			downloadLocation = filepath.Base(strings.TrimSuffix(url, urlExt))
		}
	}

	steps = append(steps,
		zen_targets.NewTarget(
			mainStepName,
			zen_targets.WithLabels(rfc.Labels),
			zen_targets.WithHashes(rfc.Hashes),
			zen_targets.WithOuts([]string{downloadLocation}),
			zen_targets.WithPassEnv(rfc.PassEnv),
			zen_targets.WithSecretEnvVars(rfc.SecretEnv),
			zen_targets.WithVisibility(rfc.Visibility),
			zen_targets.WithTargetScript("build", &zen_targets.TargetScript{
				Deps: rfc.Deps,
				Run: func(target *zen_targets.Target, runCtx *zen_targets.RuntimeContext) error {
					var interpolatedDownload string
					if id, err := target.Interpolate(target.Outs[0]); err != nil {
						return fmt.Errorf("interpolating url: %w", err)
					} else {
						interpolatedDownload = filepath.Join(target.Cwd, id)
					}

					target.Debugln("Download url: %s to %s", url, interpolatedDownload)
					req, err := http.NewRequest("GET", url, nil)
					if err != nil {
						return err
					}

					for k, v := range rfc.Headers {
						interpolHeaderVal, err := target.Interpolate(v)
						if err != nil {
							return fmt.Errorf("interpolating header %s: %w", k, err)
						}
						req.Header.Add(k, interpolHeaderVal)
					}
					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						return err
					}
					defer resp.Body.Close()

					f, err := os.OpenFile(interpolatedDownload, os.O_CREATE|os.O_WRONLY, 0644)
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
	)

	return steps, nil
}
