package handlers

import (
	"net/http"
	"testing"

	"argus-backend/internal/auth"
	"argus-backend/internal/store"

	"github.com/google/uuid"
)

func testMemoryStoreUser(t *testing.T, st *store.MemoryStore) store.User {
	t.Helper()
	u := store.User{Email: "handler+" + t.Name() + "@example.com", PasswordHash: "x"}
	if err := st.CreateUser(u); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	out, ok := st.GetUserByEmail(u.Email)
	if !ok {
		t.Fatal("missing user")
	}
	return out
}

func withAuthUser(r *http.Request, u store.User) *http.Request {
	return r.WithContext(auth.ContextWithUser(r.Context(), u))
}

// propertyStoreAuthUser creates a user for gopter callbacks (no *testing.T).
func propertyStoreAuthUser(st *store.MemoryStore) (store.User, bool) {
	u := store.User{
		Email:        uuid.New().String() + "@property.handlers",
		PasswordHash: "x",
	}
	if err := st.CreateUser(u); err != nil {
		return store.User{}, false
	}
	out, ok := st.GetUserByEmail(u.Email)
	return out, ok
}
