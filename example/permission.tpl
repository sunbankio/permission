        sessionMap,err := utils.ExtractContextValue(r.Context(), "session-user")
        if err != nil {
            httpx.WriteJson(w, http.StatusForbidden, map[string]string{"error": "session not found"})
        }

        if _, ok := sessionMap["role"]; !ok{
            httpx.WriteJson(w, http.StatusForbidden, map[string]string{"error": "invalid session"})
        }

        if sessionMap["role"].Permissions == nil {
            httpx.WriteJson(w, http.StatusForbidden, map[string]string{"error": "session permissions not found"})
        }
        utils.HasPermission(sessionMap["role"].Permissions,"%s")