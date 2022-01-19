package cell

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
)

// extractValueFromCell fetches the value at the field path out of the cell.
//
// The implementation of this function is highly insipired by how Argo Rollouts
// transforms a field path to extract a value out of a rollout object.
//
// See the below for how Argo CD treats ValueFrom/FieldPath
// https://github.com/argoproj/argo-rollouts/blob/4739bcd2d9b4910805fd9529d5afd385dbe51bbd/utils/analysis/factory.go#L32-L43
//
// See the below for the original implementation of extractValueFromRollout
// https://github.com/argoproj/argo-rollouts/blob/4739bcd2d9b4910805fd9529d5afd385dbe51bbd/utils/analysis/factory.go#L231
func extractValueFromCell(r *okrav1alpha1.Cell, path string) (string, error) {
	j, _ := json.Marshal(r)
	m := interface{}(nil)
	json.Unmarshal(j, &m)
	sections := regexp.MustCompile("[\\.\\[\\]]+").Split(path, -1)
	for _, section := range sections {
		if section == "" {
			continue // if path ends with a separator char, Split returns an empty last section
		}

		if asArray, ok := m.([]interface{}); ok {
			if i, err := strconv.Atoi(section); err != nil {
				return "", fmt.Errorf("invalid index '%s'", section)
			} else if i >= len(asArray) {
				return "", fmt.Errorf("index %d out of range", i)
			} else {
				m = asArray[i]
			}
		} else if asMap, ok := m.(map[string]interface{}); ok {
			m = asMap[section]
		} else {
			return "", fmt.Errorf("invalid path %s in cell", path)
		}
	}

	if m == nil {
		return "", fmt.Errorf("invalid path %s in cell", path)
	}

	var isArray, isMap bool
	_, isArray = m.([]interface{})
	_, isMap = m.(map[string]interface{})
	if isArray || isMap {
		return "", fmt.Errorf("path %s in cell must terminate in a primitive value", path)
	}

	return fmt.Sprintf("%v", m), nil
}
