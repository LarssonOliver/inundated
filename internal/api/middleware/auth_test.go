package middleware_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/api/middleware"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/stretchr/testify/require"
)

func TestGetCurrentUserFromContext(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name  string
		r     *http.Request
		want  model.User
		want2 bool
	}{
		{
			name: "user in context",
			r: func() *http.Request {
				req, _ := http.NewRequest("GET", "/", nil)
				user := model.User{Id: id, Name: "Test User"}
				ctx := context.WithValue(req.Context(), "user", user)
				return req.WithContext(ctx)
			}(),
			want:  model.User{Id: id, Name: "Test User"},
			want2: true,
		},
		{
			name: "no user in context",
			r: func() *http.Request {
				req, _ := http.NewRequest("GET", "/", nil)
				return req
			}(),
			want:  model.User{},
			want2: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2 := middleware.GetCurrentUserFromContext(tt.r)
			require.Equal(t, tt.want2, got2)
			require.Equal(t, tt.want, got)
		})
	}
}
