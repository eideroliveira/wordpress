package wordpress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
)

const (
	CollectionUsers      = "users"
	CollectionPosts      = "posts"
	CollectionPages      = "pages"
	CollectionMedia      = "media"
	CollectionMeta       = "meta"
	CollectionRevisions  = "revisions"
	CollectionComments   = "comments"
	CollectionTaxonomies = "taxonomies"
	CollectionTerms      = "terms"
	CollectionStatuses   = "statuses"
	CollectionTypes      = "types"
)

type GeneralError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    int    `json:"data"` // Unsure if this is consistent
}

type Options struct {
	BaseAPIURL string
	Debug      bool

	// Basic Auth
	Username string
	Password string
	// TODO: support OAuth authentication
}

type Client struct {
	httpClient *http.Client
	options    *Options
	baseURL    string
}

// Used to create a new http.Client object.
func newHTTPClient(options *Options) *http.Client {
	transport := &http.Transport{
		DisableKeepAlives: true,
		// TODO: Add other transport configurations if needed, e.g., TLS, Proxy
	}
	client := &http.Client{
		Transport: transport,
		Jar:       nil, // Explicitly nil, as gorequest did
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Apply basic auth to all redirect requests
			if options.Username != "" && options.Password != "" {
				req.SetBasicAuth(options.Username, options.Password)
			}
			if options.Debug {
				log.Printf("REDIRECT: Request to %s via %d hops", req.URL, len(via))
			}
			// Default policy: allow up to 10 redirects.
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	return client
}

func NewClient(options *Options) *Client {
	httpClient := newHTTPClient(options)
	// Basic auth will be set per request, or handled by CheckRedirect for subsequent requests.
	// Debug logging for requests will need to be handled manually if required, outside of client setup.
	return &Client{
		httpClient: httpClient,
		options:    options,
		baseURL:    options.BaseAPIURL,
	}
}

func (client *Client) Users() *UsersCollection {
	return &UsersCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionUsers),
	}
}
func (client *Client) Posts() *PostsCollection {
	return &PostsCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionPosts),
	}
}
func (client *Client) Pages() *PagesCollection {
	return &PagesCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionPages),
	}
}
func (client *Client) Media() *MediaCollection {
	return &MediaCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionMedia),
	}
}
func (client *Client) Comments() *CommentsCollection {
	return &CommentsCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionComments),
	}
}
func (client *Client) Taxonomies() *TaxonomiesCollection {
	return &TaxonomiesCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionTaxonomies),
	}
}
func (client *Client) Terms() *TermsCollection {
	return &TermsCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionTerms),
	}
}
func (client *Client) Statuses() *StatusesCollection {
	return &StatusesCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionStatuses),
	}
}
func (client *Client) Types() *TypesCollection {
	return &TypesCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionTypes),
	}
}

func (client *Client) List(url_ string, params interface{}, result interface{}) (*http.Response, []byte, error) {
	req, err := http.NewRequest("GET", url_, nil)
	if err != nil {
		return nil, nil, err
	}

	if params != nil {
		query := req.URL.Query()
		// Assuming params is a struct or map[string]...
		// This part needs a robust way to convert params to query string
		// For simplicity, let's assume params is map[string]string for now
		// A more generic solution would use reflection or a library
		if pMap, ok := params.(map[string]string); ok {
			for k, v := range pMap {
				query.Add(k, v)
			}
		} else if pVal, ok := params.(url.Values); ok {
			req.URL.RawQuery = pVal.Encode()
		} else {
			// Fallback or error if params type is not handled
			// For now, attempting to marshal and unmarshal to url.Values
			// This is inefficient and might not cover all cases.
			// Consider using a library like github.com/google/go-querystring
			jsonParams, _ := json.Marshal(params)
			var mapParams map[string]interface{}
			_ = json.Unmarshal(jsonParams, &mapParams)
			q := url.Values{}
			for k, v := range mapParams {
				q.Add(k, fmt.Sprintf("%v", v))
			}
			req.URL.RawQuery = q.Encode()
		}
		if req.URL.RawQuery == "" && params != nil { // if not already set by url.Values
			// Simplified: convert params to query string (needs improvement for complex types)
			// This is a placeholder for a more robust query parameter encoding
			q := req.URL.Query()
			jsonParams, _ := json.Marshal(params)
			var mapParams map[string]interface{}
			_ = json.Unmarshal(jsonParams, &mapParams)
			for k, v := range mapParams {
				// Handle slices and other types appropriately
				switch val := v.(type) {
				case []interface{}:
					for _, item := range val {
						q.Add(k, fmt.Sprintf("%v", item))
					}
				default:
					q.Add(k, fmt.Sprintf("%v", val))
				}
			}
			req.URL.RawQuery = q.Encode()
		}
	}

	req.Header.Set("Accept", "application/json")
	if client.options.Username != "" && client.options.Password != "" {
		req.SetBasicAuth(client.options.Username, client.options.Password)
	}

	if client.options.Debug {
		log.Printf("Request: GET %s, Params: %+v\\n", url_, params)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, body, err
	}

	if client.options.Debug {
		log.Printf("Response: %s, Body: %s\\n", resp.Status, string(body))
	}

	err = unmarshallResponse(resp, body, result)
	return resp, body, err
}
func (client *Client) Create(url string, content interface{}, result interface{}) (*http.Response, []byte, error) {
	contentVal := unpackInterfacePointer(content)
	jsonBody, err := json.Marshal(contentVal)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshalling content: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if client.options.Username != "" && client.options.Password != "" {
		req.SetBasicAuth(client.options.Username, client.options.Password)
	}

	if client.options.Debug {
		log.Printf("Request: POST %s, Body: %s\\n", url, string(jsonBody))
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, jsonBody, err // return jsonBody as it might be useful for debugging
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, body, err
	}

	if client.options.Debug {
		log.Printf("Response: %s, Body: %s\\n", resp.Status, string(body))
	}

	err = unmarshallResponse(resp, body, result)
	return resp, body, err
}
func (client *Client) Get(url string, params interface{}, result interface{}) (*http.Response, []byte, error) {
	// Similar to List, using GET
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	if params != nil {
		// Simplified: convert params to query string (needs improvement for complex types)
		q := req.URL.Query()
		jsonParams, _ := json.Marshal(params)
		var mapParams map[string]interface{}
		_ = json.Unmarshal(jsonParams, &mapParams)
		for k, v := range mapParams {
			// Handle slices and other types appropriately
			switch val := v.(type) {
			case []interface{}:
				for _, item := range val {
					q.Add(k, fmt.Sprintf("%v", item))
				}
			default:
				q.Add(k, fmt.Sprintf("%v", val))
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	req.Header.Set("Accept", "application/json")
	if client.options.Username != "" && client.options.Password != "" {
		req.SetBasicAuth(client.options.Username, client.options.Password)
	}

	if client.options.Debug {
		log.Printf("Request: GET %s, Params: %+v\\n", url, params)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, body, err
	}

	if client.options.Debug {
		log.Printf("Response: %s, Body: %s\\n", resp.Status, string(body))
	}

	err = unmarshallResponse(resp, body, result)
	return resp, body, err
}
func (client *Client) Update(url string, content interface{}, result interface{}) (*http.Response, []byte, error) {
	contentVal := unpackInterfacePointer(content)
	jsonBody, err := json.Marshal(contentVal)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshalling content: %w", err)
	}

	// WordPress API might expect PUT or POST with X-HTTP-Method-Override
	// The original code used POST with X-HTTP-Method-Override: PUT
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-HTTP-Method-Override", "PUT") // As per original logic
	if client.options.Username != "" && client.options.Password != "" {
		req.SetBasicAuth(client.options.Username, client.options.Password)
	}

	if client.options.Debug {
		log.Printf("Request: POST (Update via X-HTTP-Method-Override: PUT) %s, Body: %s\\n", url, string(jsonBody))
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, jsonBody, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, body, err
	}

	if client.options.Debug {
		log.Printf("Response: %s, Body: %s\\n", resp.Status, string(body))
	}

	err = unmarshallResponse(resp, body, result)
	return resp, body, err
}
func (client *Client) Delete(url_ string, params interface{}, result interface{}) (*http.Response, []byte, error) {
	// Original code used GET with _method=DELETE and X-HTTP-Method-Override: DELETE
	// Standard REST practice is to use the DELETE HTTP method.
	// Let's try with actual DELETE first, then consider the override if WP API requires it.
	req, err := http.NewRequest("DELETE", url_, nil)
	if err != nil {
		return nil, nil, err
	}

	if params != nil {
		q := req.URL.Query()
		// Assuming params is a struct or map[string]...
		jsonParams, _ := json.Marshal(params)
		var mapParams map[string]interface{}
		_ = json.Unmarshal(jsonParams, &mapParams)
		for k, v := range mapParams {
			q.Add(k, fmt.Sprintf("%v", v))
		}
		req.URL.RawQuery = q.Encode()
	}
	// The original code also added _method=DELETE as a query param.
	// And X-HTTP-Method-Override: DELETE header.
	// This suggests the server might not directly support DELETE method on the URL,
	// or there's a proxy/firewall issue.
	// For now, let's replicate the original behavior more closely if direct DELETE fails.
	// Sticking to original: GET with override
	// req, err = http.NewRequest("GET", url, nil) // Re-init for GET
	// if err != nil {
	// 	return nil, nil, err
	// }
	// if params != nil {
	// 	q := req.URL.Query()
	// 	jsonParams, _ := json.Marshal(params)
	// 	var mapParams map[string]interface{}
	// 	_ = json.Unmarshal(jsonParams, &mapParams)
	// 	for k, v := range mapParams {
	// 		q.Add(k, fmt.Sprintf("%v", v))
	// 	}
	//  q.Add("_method", "DELETE") // Original behavior
	// 	req.URL.RawQuery = q.Encode()
	// }
	// req.Header.Set("X-HTTP-Method-Override", "DELETE") // Original behavior

	// Let's try a proper DELETE request first. If WP API has issues, we can revert to override.
	// If the API truly needs GET with override:
	// The original code was:
	// req := client.req.Get(url).Query(params).Query("_method=DELETE")
	// req.Set("HTTP_X_HTTP_METHOD_OVERRIDE", "DELETE")
	// This means it was a GET request with these items.

	// Replicating original logic for DELETE:
	actualURL := url_
	if params != nil {
		// Convert params to query string
		// This is a simplified version. A robust solution would use a library or more detailed reflection.
		urlValues := make(url.Values)
		jsonParams, _ := json.Marshal(params) // Error handling omitted for brevity
		var mapParams map[string]interface{}
		_ = json.Unmarshal(jsonParams, &mapParams) // Error handling omitted
		for k, v := range mapParams {
			switch val := v.(type) {
			case []interface{}:
				for _, item := range val {
					urlValues.Add(k, fmt.Sprintf("%v", item))
				}
			default:
				urlValues.Add(k, fmt.Sprintf("%v", val))
			}
		}
		if len(urlValues) > 0 {
			actualURL = fmt.Sprintf("%s?%s", actualURL, urlValues.Encode())
		}
	}
	// Add _method=DELETE to the query string for the GET request
	parsedURL, _ := url.Parse(actualURL)
	query := parsedURL.Query()
	query.Add("_method", "DELETE")
	parsedURL.RawQuery = query.Encode()
	finalURL := parsedURL.String()

	req, err = http.NewRequest("GET", finalURL, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("X-HTTP-Method-Override", "DELETE")
	req.Header.Set("Accept", "application/json") // Assuming JSON response for delete status/message
	if client.options.Username != "" && client.options.Password != "" {
		req.SetBasicAuth(client.options.Username, client.options.Password)
	}

	if client.options.Debug {
		log.Printf("Request: GET (Delete via X-HTTP-Method-Override) %s, Params: %+v\\n", finalURL, params)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, nil, err // No body to return on request error
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, body, err
	}

	if client.options.Debug {
		log.Printf("Response: %s, Body: %s\\n", resp.Status, string(body))
	}

	err = unmarshallResponse(resp, body, result) // gorequest.End() was used, implies body might not always be JSON
	return resp, body, err
}
func (client *Client) PostData(url string, content []byte, contentType string, filename string, result interface{}) (*http.Response, []byte, error) {
	// The original comment said: "// gorequest does not support POST-ing raw data"
	// net/http supports this directly.
	// The original code snippet for PostData was incomplete but started with:
	// s := client.req.Post(url)
	// buf := bytes.NewBuffer(content)
	// It seems it was trying to build a multipart request or a raw post.
	// Given `contentType` and `filename`, this is likely for file uploads (media).

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(content))
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", contentType)
	// For file uploads, Content-Disposition is often used.
	// Example: "attachment; filename=\"" + filename + "\""
	// The WordPress API might have specific header requirements for media uploads.
	// The original gorequest code did not show how filename was used with `s.Send(contentVal)`
	// or `s.SendFile(path)`.
	// If `filename` is important, it's usually part of a multipart/form-data request.
	// For a simple raw data post, filename might be used in Content-Disposition.
	if filename != "" {
		// This is a common way, but WP API might expect something else.
		req.Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	}

	req.Header.Set("Accept", "application/json") // Assuming JSON response
	if client.options.Username != "" && client.options.Password != "" {
		req.SetBasicAuth(client.options.Username, client.options.Password)
	}

	if client.options.Debug {
		log.Printf("Request: POST %s, ContentType: %s, Filename: %s, ContentLength: %d\\n", url, contentType, filename, len(content))
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, content, err // Return original content for debugging
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, body, err
	}

	if client.options.Debug {
		log.Printf("Response: %s, Body: %s\\n", resp.Status, string(body))
	}

	err = unmarshallResponse(resp, body, result)
	return resp, body, err
}

// unpackInterfacePointer helper function (from original code, slightly adapted if needed)
// This function is crucial for gorequest's .Send() which could take interface{}
// For json.Marshal, it's also good to ensure we're not passing a pointer to a pointer.
func unpackInterfacePointer(i interface{}) interface{} {
	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			// If it's a nil pointer, return nil directly
			// or handle as appropriate for marshaling (e.g., return an empty struct or map)
			// For json.Marshal, a nil pointer to a struct marshals to "null".
			return nil
		}
		// Dereference if it's a pointer
		return val.Elem().Interface()
	}
	return i
}
