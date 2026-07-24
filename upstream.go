package pluginhelper

import (
	"encoding/json"
	"strings"
)

// UpstreamInvocation is the normalized input passed to an upstream action.
// Channel.Config contains the per-channel settings declared by UpstreamType.
type UpstreamInvocation struct {
	Channel UpstreamChannel `json:"channel"`
	Request UpstreamRequest `json:"request"`
}

type UpstreamChannel struct {
	ID      uint   `json:"id"`
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Config  Values `json:"config"`
}

type UpstreamRequest struct {
	Payload map[string]any `json:"payload"`
	Stream  bool           `json:"stream"`
	Compact bool           `json:"compact"`
}

// Upstream decodes an upstream.prepare or upstream.refresh action input.
func (c *ActionContext) Upstream() (UpstreamInvocation, error) {
	var invocation UpstreamInvocation
	raw, err := json.Marshal(c.Values)
	if err != nil {
		return UpstreamInvocation{}, err
	}
	if err := json.Unmarshal(raw, &invocation); err != nil {
		return UpstreamInvocation{}, ErrorWithCode("invalid_request", "invalid upstream invocation")
	}
	if invocation.Channel.Config == nil {
		invocation.Channel.Config = Values{}
	}
	if invocation.Request.Payload == nil {
		invocation.Request.Payload = map[string]any{}
	}
	return invocation, nil
}

// UpstreamPreparedRequest is the HTTP request returned to the Community proxy.
type UpstreamPreparedRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// JSONPostRequest creates the JSON POST request shape accepted by an upstream
// action result. It always sets Content-Type to application/json.
func JSONPostRequest(url string, payload any, headers map[string]string) (UpstreamPreparedRequest, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return UpstreamPreparedRequest{}, err
	}
	resultHeaders := make(map[string]string, len(headers)+1)
	for key, value := range headers {
		resultHeaders[key] = value
	}
	resultHeaders["Content-Type"] = "application/json"
	return UpstreamPreparedRequest{Method: "POST", URL: strings.TrimSpace(url), Headers: resultHeaders, Body: string(body)}, nil
}

// UpstreamResult is the complete action result expected by Community.
// SettingsPatch is merged into global plugin settings after a successful action.
type UpstreamResult struct {
	OK            bool                     `json:"ok"`
	Request       *UpstreamPreparedRequest `json:"request,omitempty"`
	APIKey        string                   `json:"api_key,omitempty"`
	SettingsPatch map[string]any           `json:"settings_patch,omitempty"`
}

func UpstreamRequestResult(request UpstreamPreparedRequest) UpstreamResult {
	return UpstreamResult{OK: true, Request: &request}
}

func UpstreamRefreshResult() UpstreamResult {
	return UpstreamResult{OK: true}
}

func (r UpstreamResult) WithAPIKey(apiKey string) UpstreamResult {
	r.APIKey = apiKey
	return r
}

func (r UpstreamResult) WithSettingsPatch(patch map[string]any) UpstreamResult {
	r.SettingsPatch = patch
	return r
}
