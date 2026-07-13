package content

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadSource(path string) (ContentSource, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ContentSource{}, fmt.Errorf("read content source %q: %w", path, err)
	}

	var source ContentSource
	if err := json.Unmarshal(data, &source); err != nil {
		return ContentSource{}, fmt.Errorf("parse content source %q: %w", path, err)
	}
	return source, nil
}
