package files

import (
	ahoy_targets "gitlab.com/hidothealth/platform/ahoy/src/target"
)

var KnownTargets = ahoy_targets.TargetCreatorMap{
	"remote_file": RemoteFileConfig{},
	"export_file": ExportFileConfig{},
	"filegroup":   FilegroupConfig{},
	"text_file":   TextFileConfig{},
	"transform":   TransformConfig{},
}
