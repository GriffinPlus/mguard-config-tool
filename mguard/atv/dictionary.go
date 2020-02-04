package atv

import (
	"fmt"
	"strings"
)

// dictionary represents a list of KeyValuePairs that maps keys to values.
// The dictionary preserves the order of of the key-value-pairs.
type dictionary []keyValuePair

// keyValuePair represents a pair of two strings.
type keyValuePair struct {
	Key   string `@Ident "="`
	Value string `@String`
}

// Add adds a new item to the dictionary.
func (dict *dictionary) Add(key, value string) error {
	if dict.ContainsKey(key) {
		return fmt.Errorf("Key '%s' is already in the dictionary", key)
	}
	*dict = append(*dict, keyValuePair{Key: key, Value: value})
	return nil
}

// ContainsKey checks whether the specified key is in the dictionary.
func (dict *dictionary) ContainsKey(key string) bool {
	for _, kvp := range *dict {
		if kvp.Key == key {
			return true
		}
	}
	return false
}

// TryGet returns the value associated with the specified key.
// If the key does not exist, nil is returned.
func (dict *dictionary) TryGet(key string, value *string) bool {
	for _, kvp := range *dict {
		if kvp.Key == key {
			*value = kvp.Value
			return true
		}
	}
	return false
}

// Set sets the item with the specified key to the specified value.
// The item is created if it does not exist, yet.
func (dict *dictionary) Set(key string, value string) {

	// set value if the item exists already
	for _, kvp := range *dict {
		if kvp.Key == key {
			kvp.Value = value
			return
		}
	}

	// add new item
	*dict = append(*dict, keyValuePair{Key: key, Value: value})
}

// Remove removes the item with the specified key.
func (dict *dictionary) Remove(key string) bool {
	for i, kvp := range *dict {
		if kvp.Key == key {
			a := []keyValuePair(*dict)
			copy(a[i:], a[i+1:])
			a[len(a)-1] = keyValuePair{}
			*dict = a[:len(a)-1]
			return true
		}
	}
	return false
}

// String gets the string representation of the dictionary.
func (dict *dictionary) String() string {
	builder := strings.Builder{}
	for _, kvp := range *dict {
		s := fmt.Sprintf("Key: '%s' => Value: '%s'\n", kvp.Key, kvp.Value)
		builder.WriteString(s)
	}
	return strings.TrimSpace(builder.String())
}
