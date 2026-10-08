package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

var envName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Environment extracts environment-variable references from Codex/Pi JSON
// configuration and resolves them using the host environment. Literal secrets
// remain in the read-only config mount and are not copied into this map.
func Environment(harness string, files []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read %s config: %w", harness, err)
		}
		var value any
		if strings.HasSuffix(file, ".toml") {
			var parsed map[string]any
			err = toml.Unmarshal(data, &parsed)
			value = parsed
		} else {
			err = json.Unmarshal(data, &value)
		}
		if err != nil {
			return nil, fmt.Errorf("parse %s config %s: %w", harness, file, err)
		}
		collect(value, strings.HasSuffix(file, "auth.json"), result)
	}
	return result, nil
}

func collect(value any, allowBareName bool, result map[string]string) {
	switch value := value.(type) {
	case map[string]any:
		for key, child := range value {
			if s, ok := child.(string); ok && key == "env_key" {
				addReference("$"+s, false, result)
			} else if s, ok := child.(string); ok && isSecretField(key) {
				addReference(s, allowBareName, result)
			} else {
				collect(child, allowBareName, result)
			}
		}
	case []any:
		for _, child := range value {
			collect(child, allowBareName, result)
		}
	}
}

func isSecretField(key string) bool {
	key = strings.ToLower(key)
	return key == "key" || key == "apikey" || key == "api_key" || key == "x-api-key" || key == "authorization"
}

func addReference(value string, allowBareName bool, result map[string]string) {
	name := value
	if strings.HasPrefix(name, "${") && strings.HasSuffix(name, "}") {
		name = name[2 : len(name)-1]
	} else if strings.HasPrefix(name, "$") {
		name = name[1:]
	}
	// Pi auth.json permits an environment variable name as the key value.
	// models.json uses $NAME/${NAME}; literal API keys are left in the mount.
	if !envName.MatchString(name) || (!allowBareName && !strings.HasPrefix(value, "$") && !strings.HasPrefix(value, "${")) {
		return
	}
	if resolved, ok := os.LookupEnv(name); ok {
		result[name] = resolved
	}
}
