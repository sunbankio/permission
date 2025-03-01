        sessionMap,err := utils.ExtractContextValue(r.Context(), "session-user")
        if err != nil {
            response.Response(w, nil, apierror.ErrSessionNotFound)
            return 
        }

        if _, ok := sessionMap["role"]; !ok{
            response.Response(w, nil, apierror.ErrInvalidSession)
            return
        }

        if sessionMap["role"].Permissions == nil {
            response.Response(w, nil, apierror.ErrPermissionsEmpty)
            return
        }

        if !utils.HasPermission(userContext.Role.Permissions, "%s") {
			response.Response(w, nil, apierror.ErrNotAuthorized)
			return
		}