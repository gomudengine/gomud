package buffs

import (
	"io/fs"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/fileloader"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"gopkg.in/yaml.v2"
)

var (
	pluginFlagFileSystems []fileloader.ReadableGroupFS
)

// RegisterFlagFS registers a plugin file system to be searched when loading
// buff flag data files. Must be called before LoadFlagDataFiles().
func RegisterFlagFS(f ...fileloader.ReadableGroupFS) {
	pluginFlagFileSystems = append(pluginFlagFileSystems, f...)
}

// loadPluginFlags walks every sub-filesystem of every registered plugin FS,
// reading flag YAML files from a "buffs-flags/" prefix and merging them into
// dst. Disk-loaded flags take precedence: a plugin flag with a duplicate id is
// logged and skipped.
func loadPluginFlags(dst map[string]*FlagSpec) {
	for _, groupFS := range pluginFlagFileSystems {
		for subFS := range groupFS.AllFileSubSystems {
			loadFlagsFromFS(subFS, dst)
		}
	}
}

func loadFlagsFromFS(subFS fs.ReadFileFS, dst map[string]*FlagSpec) {
	if pl, ok := subFS.(fileloader.PathLister); ok {
		for _, path := range pl.KnownPaths() {
			if !strings.HasPrefix(path, `buffs-flags/`) || !strings.HasSuffix(path, `.yaml`) {
				continue
			}
			loadFlagFileFromFS(subFS, path, dst)
		}
		return
	}

	_ = fs.WalkDir(subFS, `buffs-flags`, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, `.yaml`) {
			return nil
		}
		loadFlagFileFromFS(subFS, path, dst)
		return nil
	})
}

func loadFlagFileFromFS(subFS fs.ReadFileFS, path string, dst map[string]*FlagSpec) {
	b, err := subFS.ReadFile(path)
	if err != nil {
		mudlog.Error("buffs.loadFlagsFromFS", "path", path, "error", err)
		return
	}

	var spec FlagSpec
	if err := yaml.Unmarshal(b, &spec); err != nil {
		mudlog.Error("buffs.loadFlagsFromFS", "path", path, "error", err)
		return
	}

	if err := spec.Validate(); err != nil {
		mudlog.Error("buffs.loadFlagsFromFS", "path", path, "error", err)
		return
	}

	if !strings.HasSuffix(path, spec.Filepath()) {
		mudlog.Error("buffs.loadFlagsFromFS", "path", path, "expected suffix", spec.Filepath(), "error", "filepath mismatch")
		return
	}

	if _, exists := dst[spec.Flag]; exists {
		mudlog.Error("buffs.loadFlagsFromFS", "flag", spec.Flag, "path", path, "error", "duplicate flag id")
		return
	}

	dst[spec.Flag] = &spec
}
