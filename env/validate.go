package env

import (
	"fmt"
	"log"
)

func Assign_uint64(k string, v any, target *uint64) error {
	switch t := v.(type) {
	case float64:
		*target = uint64(t)
	case uint64:
		*target = t
	default:
		return Fatal_log(Type_error(k, v))
	}
	return nil
}

func Read_only(k string) error {
	return Fatal_log(fmt.Errorf("Can not change %s after init", k))
}

func Extract_lang(p map[string]any) (string, error){
	lang, ok := p["lang"]
	if !ok {
		return "", nil
	}
	s, ok := lang.(string)
	if !ok {
		return "", Fatal_log(Type_error("lang", lang))
	}
	delete(p, "lang")
	return s, nil
}

func Fatal_log(err error) error {
	log.Printf("Env: %v", err)
	return err
}

func Key_error(key string) error {
	return fmt.Errorf("Invalid env key: %s", key)
}

func Type_error(key string, value any) error {
	return fmt.Errorf("Invalid env key type: %s (%T)", key, value)
}