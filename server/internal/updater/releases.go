// Package updater checks a fixed GitHub release source and installs verified
// native binaries through a separate helper process after graceful shutdown.
package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const Protocol = 1
const maxBinarySize int64 = 512 << 20

type Build struct {
	Codename       string `json:"codename,omitempty"`
	Repository     string `json:"repository,omitempty"`
	Version        string `json:"version"`
	Commit         string `json:"commit"`
	Date           string `json:"date"`
	OS             string `json:"os"`
	Arch           string `json:"arch"`
	UpdateProtocol int    `json:"updateProtocol"`
}

type Artifact struct {
	Name   string `json:"name"`
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	Kind   string `json:"kind"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

type manifest struct {
	Codename       string     `json:"codename,omitempty"`
	SchemaVersion  int        `json:"schemaVersion"`
	UpdateProtocol int        `json:"updateProtocol"`
	Version        string     `json:"version"`
	Commit         string     `json:"commit"`
	Artifacts      []Artifact `json:"artifacts"`
}

type Release struct {
	ID          int64  `json:"id"`
	Tag         string `json:"tag"`
	Codename    string `json:"codename,omitempty"`
	Prerelease  bool   `json:"prerelease"`
	Installable bool   `json:"installable"`
	Reason      string `json:"reason,omitempty"`
	Relation    string `json:"relation,omitempty"`
	Version     string `json:"version"`
	Name        string `json:"name"`
	Notes       string `json:"notes"`
	URL         string `json:"url"`
	PublishedAt string `json:"publishedAt"`
	Size        int64  `json:"size"`
	artifact    Artifact
	assetID     int64
	commit      string
}

var repoRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,99}/[A-Za-z0-9][A-Za-z0-9_.-]{0,99}$`)
var hashRE = regexp.MustCompile(`^[0-9a-f]{64}$`)
var commitRE = regexp.MustCompile(`^[0-9a-f]{40}$`)

func newClient() *http.Client {
	return &http.Client{Timeout: 15 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("too many redirects")
		}
		req.Header.Del("Authorization")
		host := req.URL.Hostname()
		if req.URL.Scheme != "https" || (host != "github.com" && host != "api.github.com" && !strings.HasSuffix(host, ".githubusercontent.com")) {
			return errors.New("untrusted release redirect")
		}
		return nil
	}}
}

func (m *Manager) request(ctx context.Context, endpoint string, binary bool) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.apiBase+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Myself/"+m.opts.Build.Version)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Accept", "application/vnd.github+json")
	if binary {
		req.Header.Set("Accept", "application/octet-stream")
	}
	if m.token != "" {
		req.Header.Set("Authorization", "Bearer "+m.token)
	}
	response, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		if response.StatusCode == http.StatusNotFound {
			return nil, errors.New("repository_unavailable")
		}
		return nil, fmt.Errorf("github_http_%d", response.StatusCode)
	}
	return response, nil
}

func readJSON(response *http.Response, target any) error {
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, (8<<20)+1))
	if err != nil {
		return err
	}
	if len(data) > 8<<20 {
		return errors.New("release metadata too large")
	}
	return json.Unmarshal(data, target)
}

func (m *Manager) assetPath(id int64) string {
	return "/repos/" + m.opts.Repository + "/releases/assets/" + strconv.FormatInt(id, 10)
}
