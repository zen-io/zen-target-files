package files

import (
	zen_targets "github.com/zen-io/zen-core/target"
)

var KnownTargets = zen_targets.TargetCreatorMap{
	"remote_file": RemoteFileConfig{},
	"export_file": ExportFileConfig{},
	"filegroup":   FilegroupConfig{},
	"text_file":   TextFileConfig{},
	"transform":   TransformConfig{},
}
