package routes

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/akhmadzaki/go-backend/config"
)

func TestSwaggerUI(t *testing.T) {
	cfg := config.Load()
	router := RegisterRoutes(&cfg, nil)

	tt := []struct {
		path string
		want string
	}{
		{"/swagger/index.html", "Swagger UI"},
		{"/swagger/doc.json", "\"swagger\": \"2.0\""},
		{"/swagger/", "Swagger UI"},
	}

	for _, tc := range tt {
		req := httptest.NewRequest("GET", tc.path, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if tc.path == "/swagger/" && rr.Code == 301 {
			// Expected: root swagger path redirects to index.html
			continue
		}
		if rr.Code != 200 {
			t.Fatalf("%s: got status %d, want 200", tc.path, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), tc.want) {
			t.Fatalf("%s: body missing %q", tc.path, tc.want)
		}
	}
}
