package file

import "strings"

// parseFile parses the contents of a NIML file and returns a map with the data.
// NIML format:
//   - (topic) - topic/section declaration
//   - / key = "value" - key-value pair declaration
//   - ;; comment - single-line comment
//
// Comments can be either standalone lines or inline after values.
func ParseFile(data string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	lines := strings.Split(data, "\n")

	var currentTopic string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, ";;") {
			continue
		}

		if idx := strings.Index(line, ";;"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "(") && strings.HasSuffix(line, ")") {
			currentTopic = strings.Trim(line, "()")
			if _, exists := result[currentTopic]; !exists {
				result[currentTopic] = make(map[string]interface{})
			}
			continue
		}

		if strings.HasPrefix(line, "/") {
			line = strings.TrimPrefix(line, "/")
			line = strings.TrimSpace(line)

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, `"`)

			if currentTopic != "" {
				if topicMap, ok := result[currentTopic].(map[string]interface{}); ok {
					topicMap[key] = value
				}
			} else {
				result[key] = value
			}
		}
	}

	return result, nil
}
