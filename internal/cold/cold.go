package cold

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

var client = &http.Client{}

type Options struct {
	TargetUrl     string
	DataDir       string
	IgnoreHeaders []string
}

type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

type Response struct {
	Status      int               `json:"status"`
	Headers     map[string]string `json:"headers"`
	ContentType string            `json:"contentType"`
	Body        []byte            `json:"body"`
}

func NewRequest(r *http.Request) (*Request, error) {
	cr := Request{
		Method:  r.Method,
		URL:     r.URL.String(),
		Headers: map[string]string{},
		Body:    []byte{},
	}

	for k, v := range r.Header {
		cr.Headers[k] = v[0]
	}

	bs, err := io.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}

	cr.Body = bs

	return &cr, nil
}

func (r *Request) GetHash(opts *Options) (string, error) {
	rr := Request{
		Method:  r.Method,
		URL:     r.URL,
		Headers: map[string]string{},
		Body:    r.Body,
	}

	for k, v := range r.Headers {
		ignored := false

		for _, h := range opts.IgnoreHeaders {
			if strings.ToLower(k) == strings.ToLower(h) {
				fmt.Printf("ignoring header %s\n", k)

				delete(r.Headers, k)

				ignored = true
			}
		}

		if !ignored {
			rr.Headers[k] = v
		}
	}

	bs, err := json.Marshal(r)

	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(bs)

	return hex.EncodeToString(h.Sum(nil)), nil
}

func fromFile(fp string) (*Response, error) {
	fmt.Printf("serving from file %s\n", fp)

	cresp := Response{Headers: map[string]string{}, Body: []byte{}}

	bs, err := os.ReadFile(fp)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bs, &cresp)

	if err != nil {
		return nil, err
	}

	return &cresp, nil
}

func fromRequest(r *Request, targetUrl string, fp string) (*Response, error) {
	fmt.Printf("serving fresh %s\n", targetUrl+r.URL)

	req, err := http.NewRequest(r.Method, targetUrl+r.URL, bytes.NewReader(r.Body))

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	cresp := Response{
		Status:      resp.StatusCode,
		Headers:     map[string]string{},
		ContentType: resp.Header.Get("Content-Type"),
		Body:        bs,
	}

	for k, v := range resp.Header {
		cresp.Headers[k] = v[0]
	}

	bs, err = json.Marshal(cresp)

	if err != nil {
		return nil, err
	}

	err = os.WriteFile(fp, bs, 0666)

	if err != nil {
		return nil, err
	}

	return &cresp, nil
}

func (r *Request) ToResponse(opts *Options) (*Response, error) {
	h, err := r.GetHash(opts)

	if err != nil {
		return nil, err
	}

	fp := path.Join(opts.DataDir, h+".json")

	if _, err := os.Stat(fp); err == nil {
		return fromFile(fp)
	} else if errors.Is(err, os.ErrNotExist) {
		return fromRequest(r, opts.TargetUrl, fp)
	}

	return nil, err
}
