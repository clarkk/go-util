package obj

import (
	//"fmt"
	"sort"
	//"reflect"
)

func Map_sort_key[V any](m map[int]V, reverse bool) []V {
	length := len(m)
	keys := make([]int, length)
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Slice(keys, func(a, b int) bool {
		if reverse {
			return keys[a] > keys[b]
		}
		return keys[a] < keys[b]
	})
	
	list := make([]V, length)
	i = 0
	for _, k := range keys {
		list[i] = m[k]
		i++
	}
	return list
}

/*func Map_struct_types(input any) (map[string]reflect.Type, error) {
	list := map[string]reflect.Type{}
	struct_type := reflect.TypeOf(input).Elem()
	if struct_type.Kind() != reflect.Struct {
		return list, fmt.Errorf("Struct must be a struct")
	}
	struct_val := reflect.ValueOf(input).Elem().Type()
	for i := range struct_val.NumField() {
		field := struct_val.Field(i)
		list[field.Name] = field.Type
	}
	return list, nil
}*/