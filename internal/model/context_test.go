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

