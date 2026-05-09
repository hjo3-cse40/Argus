package handlers

import (
	"net/http"

	"argus-backend/internal/auth"
)

const testUserID = "test-user-id"

func withTestAuth(req *http.Request) *http.Request {
	ctx := auth.WithUserContext(req.Context(), &auth.Claims{UserID: testUserID})
	return req.WithContext(ctx)
}
