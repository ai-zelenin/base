// Package influx contains influx client, configs and helpers
package influx

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/influxdb/models"
	"golang.org/x/net/http2"
)

type Point struct {
	Name   string
	Tags   map[string]string
	Fields map[string]interface{}
	Time   time.Time
}

type Row models.Row

// NewInfluxDBClient returns a new Client from the provided config.
// Client is safe for concurrent use by multiple goroutines.
func NewInfluxDBClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config for influx client == nil")
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DisableCompression:    false,
		MaxConnsPerHost:       1000,
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   1000,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		IdleConnTimeout:       config.IdleConnTimeout,
	}

	err := http2.ConfigureTransport(tr)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}

	if config.UserAgent == "" {
		config.UserAgent = "Client"
	}

	writeURL, err := makeWriteURL(
		u,
		config.Database,
		config.RetentionPolicy,
		config.Consistency)
	if err != nil {
		return nil, err
	}

	queryURL, err := makeQueryURL(u)
	if err != nil {
		return nil, err
	}
	ic := &Client{
		baseURL:  *u,
		queryURL: queryURL,
		writeURL: writeURL,
		Cfg:      config,
		httpClient: &http.Client{
			Timeout:   config.Timeout,
			Transport: tr,
		},
		transport: tr,
		headers:   make(map[string]string),
	}

	ic.headers["User-Agent"] = ic.Cfg.UserAgent

	return ic, nil
}

// Client is safe for concurrent use as the fields are all read-only
// once the Client is instantiated.
type Client struct {
	// N.B - if baseURL.UserInfo is accessed in future modifications to the
	// methods on Client, you will need to synchronize access to baseURL.
	baseURL    url.URL
	writeURL   string
	queryURL   string
	Cfg        *Config
	httpClient *http.Client
	transport  *http.Transport
	headers    map[string]string
}

// Ping will check to see if the server is up with an optional timeout on waiting for leader.
// Ping returns how long the request took, the version of the server it connected to, and an error if one occurred.
func (ic *Client) Ping(timeout time.Duration) (time.Duration, string, error) {
	now := time.Now()

	u := ic.baseURL
	u.Path = path.Join(u.Path, "ping")

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return 0, "", err
	}

	// Add headers
	if ic.Cfg.Username != "" || ic.Cfg.Password != "" {
		req.SetBasicAuth(ic.Cfg.Username, ic.Cfg.Password)
	}
	for header, value := range ic.headers {
		req.Header.Set(header, value)
	}

	if timeout > 0 {
		params := req.URL.Query()
		params.Set("wait_for_leader", fmt.Sprintf("%.0fs", timeout.Seconds()))
		req.URL.RawQuery = params.Encode()
	}

	resp, err := ic.httpClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}
	err = resp.Body.Close()
	if err != nil {
		return 0, "", err
	}

	if resp.StatusCode != http.StatusNoContent {
		var err = fmt.Errorf(string(body))
		return 0, "", err
	}

	version := resp.Header.Get("X-Influxdb-Version")
	return time.Since(now), version, nil
}

// SendQuery query in default db
func (ic *Client) SendQuery(ctx context.Context, q string) ([]Row, error) {
	response, err := ic.Query(ctx, influx.Query{
		Command:    q,
		Parameters: make(map[string]interface{}),
	})
	if err != nil {
		return nil, err
	}
	rowSet := make([]Row, 0)
	for _, result := range response.Results {
		for _, row := range result.Series {
			rowSet = append(rowSet, Row(row))
		}
	}
	return rowSet, nil
}

// Query sends a command to the server and returns the Response.
func (ic *Client) Query(ctx context.Context, q influx.Query) (*influx.Response, error) {
	req, err := ic.createDefaultRequest(q)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept-Encoding", "gzip")
	params := req.URL.Query()
	if q.Chunked {
		params.Set("chunked", "true")
		if q.ChunkSize > 0 {
			params.Set("chunk_size", strconv.Itoa(q.ChunkSize))
		}
		req.URL.RawQuery = params.Encode()
	}
	resp, err := ic.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err = checkResponse(resp); err != nil {
		return nil, err
	}

	var body io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		body, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer body.Close()
	default:
		body = resp.Body
	}

	var response influx.Response
	if q.Chunked {
		cr := influx.NewChunkedResponse(body)
		for {
			r, err := cr.NextResponse()
			if err != nil {
				if err == io.EOF {
					break
				}
				if err.Error() == "" {
					return nil, fmt.Errorf("unknown error on ChunkedResponse.NextResponse()")
				}
				// If we got an error while decoding the response, send that back.
				return nil, err
			}
			if r == nil {
				break
			}
			response.Results = append(response.Results, r.Results...)
			if r.Err != "" {
				response.Err = r.Err
				break
			}
		}
	} else {
		dec := json.NewDecoder(body)
		dec.UseNumber()
		decErr := dec.Decode(&response)

		// ignore this error if we got an invalid status code
		if decErr != nil && decErr.Error() == "EOF" && resp.StatusCode != http.StatusOK {
			decErr = nil
		}
		// If we got a valid decode error, send that back
		if decErr != nil {
			return nil, fmt.Errorf("unable to decode json: received status code %d err: %s", resp.StatusCode, decErr)
		}
	}

	// If we don't have an error in our json response, and didn't get statusOK
	// then send back an error
	if resp.StatusCode != http.StatusOK && response.Error() == nil {
		return &response, fmt.Errorf("received status code %d from server", resp.StatusCode)
	}
	return &response, nil
}

func (ic *Client) Write(bp influx.BatchPoints) error {
	var requestBuffer bytes.Buffer

	for _, p := range bp.Points() {
		if p == nil {
			continue
		}
		_, err := requestBuffer.WriteString(p.PrecisionString(bp.Precision()))
		if err != nil {
			return err
		}

		if err := requestBuffer.WriteByte('\n'); err != nil {
			return err
		}
	}

	var data io.Reader
	if ic.Cfg.ContentEncoding == "gzip" {
		var gzBuffer bytes.Buffer
		gzw := gzip.NewWriter(&gzBuffer)
		_, err := io.Copy(gzw, &requestBuffer)
		if err != nil {
			return err
		}
		gzw.Close()

		data = &gzBuffer
	} else {
		data = &requestBuffer
	}

	// Create new request
	req, err := http.NewRequest("POST", ic.writeURL, data)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	if ic.Cfg.ContentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}

	// Add headers
	if ic.Cfg.Username != "" || ic.Cfg.Password != "" {
		req.SetBasicAuth(ic.Cfg.Username, ic.Cfg.Password)
	}
	for header, value := range ic.headers {
		req.Header.Set(header, value)
	}

	resp, err := ic.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var err = fmt.Errorf(string(body))
		return err
	}

	return nil
}

func (ic *Client) WritePoints(points []Point) error {
	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:         ic.Cfg.Database,
		Precision:        ic.Cfg.Precision,
		RetentionPolicy:  ic.Cfg.RetentionPolicy,
		WriteConsistency: ic.Cfg.Consistency,
	})
	if err != nil {
		return err
	}

	for _, point := range points {
		p, err := influx.NewPoint(point.Name, point.Tags, point.Fields, point.Time)
		if err != nil {
			return err
		}
		bp.AddPoint(p)
	}
	return ic.Write(bp)
}

// Close releases the Client's resources.
func (ic *Client) Close() error {
	ic.transport.CloseIdleConnections()
	return nil
}

func (ic *Client) createDefaultRequest(q influx.Query) (*http.Request, error) {
	u := ic.baseURL
	u.Path = path.Join(u.Path, "query")

	jsonParameters, err := json.Marshal(q.Parameters)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "")
	req.Header.Set("User-Agent", ic.Cfg.UserAgent)

	if ic.Cfg.Username != "" {
		req.SetBasicAuth(ic.Cfg.Username, ic.Cfg.Password)
	}

	params := req.URL.Query()
	params.Set("q", q.Command)

	if q.Database == "" {
		params.Set("db", ic.Cfg.Database)
	}

	if q.RetentionPolicy == "" {
		params.Set("rp", ic.Cfg.RetentionPolicy)
	}
	params.Set("params", string(jsonParameters))

	if q.Precision == "" {
		params.Set("epoch", ic.Cfg.Precision)
	}
	req.URL.RawQuery = params.Encode()

	return req, nil
}
func checkResponse(resp *http.Response) error {
	// If we lack a X-Influxdb-Version header, then we didn't get a response from influxdb
	// but instead some other service. If the error code is also a 500+ code, then some
	// downstream loadbalancer/proxy/etc had an issue and we should report that.
	if resp.Header.Get("X-Influxdb-Version") == "" && resp.StatusCode >= http.StatusInternalServerError {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil || len(body) == 0 {
			return fmt.Errorf("received status code %d from downstream server", resp.StatusCode)
		}

		return fmt.Errorf("received status code %d from downstream server, with response body: %q", resp.StatusCode, body)
	}

	// If we get an unexpected content type, then it is also not from influx direct and therefore
	// we want to know what we received and what status code was returned for debugging purposes.
	if cType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type")); cType != "application/json" {
		// Read up to 1kb of the body to help identify downstream errors and limit the impact of things
		// like downstream serving a large file
		body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1024))
		if err != nil || len(body) == 0 {
			return fmt.Errorf("expected json response, got empty body, with status: %v", resp.StatusCode)
		}

		return fmt.Errorf("expected json response, got %q, with status: %v and response body: %q", cType, resp.StatusCode, body)
	}
	return nil
}

func makeWriteURL(loc *url.URL, db, rp, consistency string) (string, error) {
	params := url.Values{}
	params.Set("db", db)

	if rp != "" {
		params.Set("rp", rp)
	}

	if consistency != "one" && consistency != "" {
		params.Set("consistency", consistency)
	}

	u := *loc
	switch u.Scheme {
	case "unix":
		u.Scheme = "http"
		u.Host = "127.0.0.1"
		u.Path = "/write"
	case "http", "https":
		u.Path = path.Join(u.Path, "write")
	default:
		return "", fmt.Errorf("unsupported scheme: %q", loc.Scheme)
	}
	u.RawQuery = params.Encode()
	return u.String(), nil
}

func makeQueryURL(loc *url.URL) (string, error) {
	u := *loc
	switch u.Scheme {
	case "unix":
		u.Scheme = "http"
		u.Host = "127.0.0.1"
		u.Path = "/query"
	case "http", "https":
		u.Path = path.Join(u.Path, "query")
	default:
		return "", fmt.Errorf("unsupported scheme: %q", loc.Scheme)
	}
	return u.String(), nil
}
