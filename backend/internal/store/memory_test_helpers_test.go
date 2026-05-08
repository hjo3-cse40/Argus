package store

import "testing"

func memoryTestUser(t *testing.T, st *MemoryStore) string {
	t.Helper()
	u := User{Email: "mem+" + t.Name() + "@example.com", PasswordHash: "x"}
	if err := st.CreateUser(u); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	out, ok := st.GetUserByEmail(u.Email)
	if !ok {
		t.Fatal("missing user")
	}
	return out.ID
}
