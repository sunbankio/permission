package utils

import (
	"context"
	"encoding/json"
	"fmt"
)

func HasPermission(userPermissions []string, requiredPermission string) bool {
	for _, permission := range userPermissions {
		if permission == requiredPermission {
			return true
		}
	}
	return false
}

func ExtractContextValue(ctx context.Context, contextKey string) (map[string]interface{}, error) {

	ctxVal := ctx.Value(contextKey)

	if ctxVal == nil {
		return nil, fmt.Errorf("context value of %s is nil", contextKey)
	}

	bytes, err := json.Marshal(ctxVal)

	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})

	json.Unmarshal(bytes, &m)

	fmt.Println(m)

	return m, nil

}
