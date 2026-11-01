package yaml_conf

import (
	"os"
	"fmt"
	"gopkg.in/yaml.v3"
)

func Load(path string, target any) error {
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