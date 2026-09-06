package model_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/stretchr/testify/require"
)

func TestGetCurrentUserFromContext(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name  string
		r     context.Context
		want  model.User
		want2 bool
	}{
		{
			name: "user in context",
			r: func() context.Context {
				req, _ := http.NewRequest("GET", "/", nil)
				user := model.User{Id: id, Name: "Test User"}
				return context.WithValue(req.Context(), model.UserContextKey, user)
			}(),
			want:  model.User{Id: id, Name: "Test User"},
			want2: true,
		},
		{
			name: "no user in context",
			r: func() context.Context {
				req, _ := http.NewRequest("GET", "/", nil)
				return req.Context()
			}(),
			want:  model.User{},
			want2: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2 := model.GetCurrentUserFromContext(tt.r)
			require.Equal(t, tt.want2, got2)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSetUserInContext(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name string // description of this test case
		user model.User
	}{
		{
			name: "set user in context",
			user: model.User{Id: id, Name: "Test User"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := model.SetUserInContext(context.Background(), tt.user)
			require.NotNil(t, got)
			retrievedUser, ok := model.GetCurrentUserFromContext(got)
			require.True(t, ok)
			require.Equal(t, tt.user, retrievedUser)
		})
	}
}

func TestGetCurrentSessionFromContext(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name    string // description of this test case
		ctx     context.Context
		session model.Session
		ok      bool
	}{
		{
			name: "session in context",
			ctx: func() context.Context {
				req, _ := http.NewRequest("GET", "/", nil)
				session := model.Session{Id: id, Sub: "sub_123"}
				return context.WithValue(req.Context(), model.SessionContextKey, session)
			}(),
			session: model.Session{Id: id, Sub: "sub_123"},
			ok:      true,
		},
		{
			name:    "no session in context",
			ctx:     context.Background(),
			session: model.Session{},
			ok:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, ok := model.GetSessionFromContext(tt.ctx)
			require.Equal(t, tt.ok, ok)
			require.Equal(t, tt.session.Id, session.Id)
		})
	}
}

func TestSetSessionInContext(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		session model.Session
	}{
		{
			name:    "set session in context",
			session: model.Session{Id: uuid.New(), Sub: "sub_123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := model.SetSessionInContext(t.Context(), tt.session)
			require.NotNil(t, got)
			retrievedSession, ok := model.GetSessionFromContext(got)
			require.True(t, ok)
			require.Equal(t, tt.session.Id, retrievedSession.Id)
		})
	}
}
