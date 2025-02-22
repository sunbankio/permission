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

func ExtractContextValue(ctx context.Context, contextKey any, destination any) error {

	ctxVal := ctx.Value(contextKey)

	if ctxVal == nil {
		return fmt.Errorf("context value of %s is nil", contextKey)
	}

	bytes, err := json.Marshal(ctxVal)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &destination); err != nil {
		return err
	}

	return nil
}
