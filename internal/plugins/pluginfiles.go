package plugins

import (
	"embed"
	"io/fs"
)

// Implements fs.ReadFileFS
type PluginFiles struct {
	fileSystem embed.FS
	filePaths  map[string]string
}

func (p PluginFiles) ReadFile(name string) ([]byte, error) {

	if embedPath, ok := p.filePaths[name]; ok {
		b, err := p.fileSystem.ReadFile(embedPath)
		if err == nil {
			return b, nil
		}
	}

	return nil, fs.ErrNotExist
}

func (p PluginFiles) Open(name string) (fs.File, error) {

	if embedPath, ok := p.filePaths[name]; ok {
		return p.fileSystem.Open(embedPath)

	}

	return nil, fs.ErrNotExist

}

func (p PluginFiles) Stat(name string) (fs.FileInfo, error) {

	if embedPath, ok := p.filePaths[name]; ok {
		return fs.Stat(p.fileSystem, embedPath)
	}

	return nil, fs.ErrNotExist

}

// KnownPaths returns all file paths registered in this plugin's file system.
func (p PluginFiles) KnownPaths() []string {
	paths := make([]string, 0, len(p.filePaths))
	for shortPath := range p.filePaths {
		paths = append(paths, shortPath)
	}
	return paths
}
