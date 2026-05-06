package history

import (
	"fmt"
	"sort"
	"strings"
)

// Tag represents a key-value label attached to a history record.
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Tags is a sortable slice of Tag.
type Tags []Tag

func (t Tags) Len() int           { return len(t) }
func (t Tags) Less(i, j int) bool { return t[i].Key < t[j].Key }
func (t Tags) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

// ParseTags parses a comma-separated list of key=value pairs into Tags.
// Example: "env=prod,region=us-east-1"
func ParseTags(raw string) (Tags, error) {
	if raw == "" {
		return nil, nil
	}
	var tags Tags
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 || kv[0] == "" {
			return nil, fmt.Errorf("invalid tag %q: must be key=value", part)
		}
		tags = append(tags, Tag{Key: strings.TrimSpace(kv[0]), Value: strings.TrimSpace(kv[1])})
	}
	sort.Sort(tags)
	return tags, nil
}

// String serialises Tags back to a canonical comma-separated string.
func (t Tags) String() string {
	parts := make([]string, len(t))
	for i, tag := range t {
		parts[i] = tag.Key + "=" + tag.Value
	}
	return strings.Join(parts, ",")
}

// Get returns the value for key, or empty string if not found.
func (t Tags) Get(key string) string {
	for _, tag := range t {
		if tag.Key == key {
			return tag.Value
		}
	}
	return ""
}
