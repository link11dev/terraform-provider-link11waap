package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListUsers_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/accounts/users", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		json.NewEncoder(w).Encode([]UserOrganization{
			{ID: "org1", Name: "Org 1", Users: []User{{ID: "u1", Email: "user@test.com"}}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	orgs, err := c.ListUsers(context.Background())
	require.NoError(t, err)
	require.Len(t, orgs, 1)
	assert.Equal(t, "org1", orgs[0].ID)
	require.Len(t, orgs[0].Users, 1)
	assert.Equal(t, "u1", orgs[0].Users[0].ID)
}

func TestListUsers_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListUsers(context.Background())
	require.Error(t, err)
}

func TestListUsers_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListUsers(context.Background())
	require.Error(t, err)
}

func TestGetUser_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/accounts/u1", r.URL.Path)
		json.NewEncoder(w).Encode(User{ID: "u1", Email: "user@test.com", ContactName: "Test User"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	user, err := c.GetUser(context.Background(), "u1")
	require.NoError(t, err)
	assert.Equal(t, "u1", user.ID)
	assert.Equal(t, "user@test.com", user.Email)
}

func TestGetUser_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetUser(context.Background(), "missing")
	require.Error(t, err)
}

func TestGetUser_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetUser(context.Background(), "u1")
	require.Error(t, err)
}

func TestCreateUser_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/accounts/users", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var req UserCreateRequest
		require.NoError(t, json.Unmarshal(body, &req))
		assert.Equal(t, "user@test.com", req.Email)
		json.NewEncoder(w).Encode(UserCreateResponse{ID: "new-user-id"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	id, err := c.CreateUser(context.Background(), &UserCreateRequest{
		Email:       "user@test.com",
		ContactName: "Test User",
		ACL:         1,
		OrgID:       "org1",
	})
	require.NoError(t, err)
	assert.Equal(t, "new-user-id", id)
}

func TestCreateUser_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"invalid email"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.CreateUser(context.Background(), &UserCreateRequest{})
	require.Error(t, err)
}

func TestCreateUser_InvalidResponseJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.CreateUser(context.Background(), &UserCreateRequest{Email: "a@b.com"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decoding response")
}

func TestUpdateUser_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/accounts/u1", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var req UserUpdateRequest
		require.NoError(t, json.Unmarshal(body, &req))
		assert.Equal(t, "Updated Name", req.ContactName)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateUser(context.Background(), "u1", &UserUpdateRequest{
		ACL:         2,
		ContactName: "Updated Name",
		Mobile:      "+1234567890",
	})
	require.NoError(t, err)
}

func TestUpdateUser_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateUser(context.Background(), "u1", &UserUpdateRequest{})
	require.Error(t, err)
}

func TestDeleteUser_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/accounts/u1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteUser(context.Background(), "u1")
	require.NoError(t, err)
}

func TestDeleteUser_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteUser(context.Background(), "u1")
	require.Error(t, err)
}
