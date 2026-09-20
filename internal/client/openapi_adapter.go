package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	vastopenapi "github.com/realnedsanders/terraform-provider-vastai/internal/client/openapi"
)

const applicationJSON = "application/json"

func newOpenAPIClient(client *VastAIClient) *vastopenapi.Client {
	generated, err := vastopenapi.NewClient(
		client.baseURL,
		vastopenapi.WithHTTPClient(client.httpClient.StandardClient()),
		vastopenapi.WithRequestEditorFn(client.editOpenAPIRequest),
	)
	if err != nil {
		// Generated client options are static and cannot currently fail. Keep the
		// constructor's existing signature while failing fast if that changes.
		panic(fmt.Sprintf("creating generated Vast.ai client: %v", err))
	}
	return generated
}

// editOpenAPIRequest preserves the authentication and request metadata used by
// the provider while generated code owns each documented method and path.
func (c *VastAIClient) editOpenAPIRequest(ctx context.Context, req *http.Request) error {
	// The combined official specification omits trailing slashes even though
	// the live API canonicalizes to them with 301 redirects. Canonicalize before
	// sending so mutations keep their method and body.
	if !strings.HasSuffix(req.URL.Path, "/") {
		req.URL.Path += "/"
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	q := req.URL.Query()
	q.Set("api_key", c.apiKey)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", applicationJSON)

	tflog.Debug(ctx, "Vast.ai API request", map[string]interface{}{
		"method": req.Method,
		"url":    sanitizeURL(req.URL),
	})
	return nil
}

func withOpenAPIQuery(params map[string]string) vastopenapi.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		query := req.URL.Query()
		for key, value := range params {
			query.Set(key, value)
		}
		req.URL.RawQuery = query.Encode()
		return nil
	}
}

func withOpenAPIJSONBody(value interface{}) (vastopenapi.RequestEditorFn, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}
	return func(_ context.Context, req *http.Request) error {
		newBody := func() io.ReadCloser {
			return io.NopCloser(bytes.NewReader(body))
		}
		req.Body = newBody()
		req.GetBody = func() (io.ReadCloser, error) { return newBody(), nil }
		req.ContentLength = int64(len(body))
		req.Header.Set("Content-Type", applicationJSON)
		return nil
	}, nil
}

func openAPIJSONBody(value interface{}) (*bytes.Reader, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}
	return bytes.NewReader(body), nil
}

func (c *VastAIClient) doOpenAPIResponse(ctx context.Context, resp *http.Response, err error, result interface{}) error {
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	return c.decodeResponse(ctx, resp, result, true)
}

func (c *VastAIClient) doOpenAPIRawResponse(ctx context.Context, resp *http.Response, err error, result interface{}) error {
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	return c.decodeResponse(ctx, resp, result, false)
}
