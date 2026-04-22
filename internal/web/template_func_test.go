package web

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestStaticAssetURL(t *testing.T) {
	t.Parallel()

	httpReq, err := http.NewRequest(http.MethodGet, "http://gomud.net/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}

	httpsReq, err := http.NewRequest(http.MethodGet, "https://gomud.net/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	httpsReq.TLS = &tls.ConnectionState{}

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
			cdnBase:   "http://files.gomud.net",
			assetPath: "/static/css/gomud.css",
			want:      "http://files.gomud.net/static/css/gomud.css",
		},
		{
			name:      "https request drops insecure cdn",
			req:       httpsReq,
			cdnBase:   "http://files.gomud.net",
			assetPath: "/static/css/gomud.css",
			want:      "/static/css/gomud.css",
		},
		{
			name:      "https request keeps secure cdn",
			req:       httpsReq,
			cdnBase:   "https://cdn.example.com/",
			assetPath: "static/js/webclient-core.js",
			want:      "https://cdn.example.com/static/js/webclient-core.js",
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
