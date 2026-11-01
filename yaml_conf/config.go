package yaml_conf

import (
	"os"
	"fmt"
	"path/filepath"
	"gopkg.in/yaml.v3"
)

func Load(path string, target any) error {
	if !filepath.IsAbs(path) {
		executable, err := os.Executable()
		if err != nil {
			return fmt.Errorf("Unable to locate executable: %w", err)
		}
		path = filepath.Join(filepath.Dir(executable), path)
	}
	
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Unable to read YAML config %q: %w", path, err)
	}
	defer f.Close()
	
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("Unable to parse YAML config %q: %w", path, err)
	}
	
	return nil
}