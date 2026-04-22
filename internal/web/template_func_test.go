package web

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
)

func TestStaticAssetURL(t *testing.T) {
	t.Parallel()

	const (
		requestHost = "request.example"
		cdnHost     = "cdn.example"
	)

	httpReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", requestHost), nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}

	httpsReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/", requestHost), nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	httpsReq.TLS = &tls.ConnectionState{}

	proxiedHTTPSReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", requestHost), nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	proxiedHTTPSReq.Header.Set("X-Forwarded-Proto", "https")

	forwardedHTTPSReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", requestHost), nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	forwardedHTTPSReq.Header.Add("Forwarded", fmt.Sprintf(`for=203.0.113.8;proto=https;host=%s`, requestHost))

	tests := []struct {
		name      string
		req       *http.Request
		cdnBase   string
		assetPath string
		want      string
	}{
		{
			name:      "local asset without cdn",
			req:       httpReq,
			cdnBase:   "",
			assetPath: "/static/css/gomud.css",
			want:      "/static/css/gomud.css",
		},
		{
			name:      "http request keeps http cdn",
			req:       httpReq,
			cdnBase:   fmt.Sprintf("http://%s", cdnHost),
			assetPath: "/static/css/gomud.css",
			want:      fmt.Sprintf("http://%s/static/css/gomud.css", cdnHost),
		},
		{
			name:      "https request drops insecure cdn",
			req:       httpsReq,
			cdnBase:   fmt.Sprintf("http://%s", cdnHost),
			assetPath: "/static/css/gomud.css",
			want:      "/static/css/gomud.css",
		},
		{
			name:      "proxied https request drops insecure cdn",
			req:       proxiedHTTPSReq,
			cdnBase:   fmt.Sprintf("http://%s", cdnHost),
			assetPath: "/static/css/gomud.css",
			want:      "/static/css/gomud.css",
		},
		{
			name:      "forwarded https request drops insecure cdn",
			req:       forwardedHTTPSReq,
			cdnBase:   fmt.Sprintf("http://%s", cdnHost),
			assetPath: "/static/css/gomud.css",
			want:      "/static/css/gomud.css",
		},
		{
			name:      "https request keeps secure cdn",
			req:       httpsReq,
			cdnBase:   fmt.Sprintf("https://%s/", cdnHost),
			assetPath: "static/js/webclient-core.js",
			want:      fmt.Sprintf("https://%s/static/js/webclient-core.js", cdnHost),
		},
		{
			name:      "invalid cdn falls back local",
			req:       httpReq,
			cdnBase:   "://bad url",
			assetPath: "/static/images/web_bg.png",
			want:      "/static/images/web_bg.png",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := staticAssetURL(tt.req, tt.cdnBase, tt.assetPath); got != tt.want {
				t.Fatalf("staticAssetURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
