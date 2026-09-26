package common

import (
	"net/http"
	"testing"
)

func TestSessionIDFromRequest(t *testing.T) {
	tests := []struct {
		name      string
		authorize string
		cookie    string
		want      string
		wantBearer bool
		wantError bool
	}{
		{name: "bearer token", authorize: "Bearer mobile-session", cookie: "web-session", want: "mobile-session", wantBearer: true},
		{name: "cookie fallback", cookie: "web-session", want: "web-session"},
		{name: "reject invalid authorization", authorize: "Basic credentials", cookie: "web-session", wantError: true},
		{name: "missing credentials", wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodPost, "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			if test.authorize != "" {
				request.Header.Set("Authorization", test.authorize)
			}
			if test.cookie != "" {
				request.AddCookie(&http.Cookie{Name: "ss", Value: test.cookie})
			}

			got, gotBearer, err := sessionIDFromRequest(request)
			if (err != nil) != test.wantError {
				t.Fatalf("sessionIDFromRequest() error = %v, wantError %v", err, test.wantError)
			}
			if got != test.want {
				t.Errorf("sessionIDFromRequest() = %q, want %q", got, test.want)
			}
			if gotBearer != test.wantBearer {
				t.Errorf("sessionIDFromRequest() bearer = %v, want %v", gotBearer, test.wantBearer)
			}
		})
	}
}