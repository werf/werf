package externalref

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func mockResolver() (http.Handler, *int) {
	calls := new(int)

	type respDef struct {
		status int
		body   string
	}

	respMap := map[string]respDef{
		"pkg:npm/lodash@4.17.21": {
			status: http.StatusOK,
			body:   `{"purl":"pkg:npm/lodash@4.17.21","purl_requested":"pkg:npm/lodash@4.17.21","url":"https://github.com/lodash/lodash","kind":"vcs","confirmed":true,"status":"confirmed","confidence":0.9,"provider":"libraries.io","resolution":"database","sources":[{"kind":"vcs","meta":{"http_status":200,"request_url":"https://libraries.io/api/npm/lodash"},"provider":"libraries.io","picked_url":"https://github.com/lodash/lodash"}]}`,
		},
		"pkg:npm/express@4.18.2": {
			status: http.StatusOK,
			body:   `{"purl":"pkg:npm/express@4.18.2","purl_requested":"pkg:npm/express@4.18.2","url":"https://github.com/expressjs/express","kind":"vcs","confirmed":true,"status":"confirmed","confidence":0.9,"provider":"libraries.io","resolution":"database"}`,
		},
		"pkg:npm/react@18.2.0": {
			status: http.StatusOK,
			body:   `{"purl":"pkg:npm/react@18.2.0","purl_requested":"pkg:npm/react@18.2.0","url":"https://github.com/facebook/react","kind":"vcs","confirmed":true,"status":"confirmed","confidence":0.9,"provider":"libraries.io","resolution":"database"}`,
		},
		"pkg:npm/empty-url-pkg@1.0.0": {
			status: http.StatusOK,
			body:   `{"purl":"pkg:npm/empty-url-pkg@1.0.0","purl_requested":"pkg:npm/empty-url-pkg@1.0.0","url":"","kind":"vcs","confirmed":false,"status":"unresolved"}`,
		},
		"pkg:npm/bad-json@1.0.0": {
			status: http.StatusOK,
			body:   `{invalid json`,
		},
		"pkg:npm/server-error@1.0.0": {
			status: http.StatusInternalServerError,
			body:   `{"error":"internal error"}`,
		},
		"pkg:npm/unknown@0.0.0": {
			status: http.StatusNotFound,
			body:   `{"error":"not found"}`,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*calls++

		if r.URL.Path != "/api/v1/resolve" {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
			return
		}

		purl := r.URL.Query().Get("purl")
		if purl == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing purl"})
			return
		}

		def, ok := respMap[purl]
		if !ok {
			for pattern, pd := range respMap {
				if strings.HasPrefix(purl, strings.TrimSuffix(pattern, "@*")) {
					def = pd
					ok = true
					break
				}
			}
		}

		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("no resolution for %s", purl)})
			return
		}

		if def.status != http.StatusOK {
			w.WriteHeader(def.status)
			_, _ = fmt.Fprint(w, def.body)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, def.body)
	})

	return handler, calls
}
