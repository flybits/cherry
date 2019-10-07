package step

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	netURL "net/url"
)

const (
	githubAcceptType = "application/vnd.github.v3+json"
	githubUserAgent  = "cherry"

	// GitHubURL is the BaseURL for GitHub.
	GitHubURL = "https://github.com"

	// GitHubAPIURL is the BaseURL for GitHub API.
	GitHubAPIURL = "https://api.github.com"
)

type (
	// GitHubReleaseData is used for creating or modifying a release.
	GitHubReleaseData struct {
		Name       string `json:"name"`
		TagName    string `json:"tag_name"`
		Target     string `json:"target_commitish"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
		Body       string `json:"body"`
	}

	// GitHubRelease represents a GitHub release.
	GitHubRelease struct {
		ID         int           `json:"id"`
		Name       string        `json:"name"`
		TagName    string        `json:"tag_name"`
		Target     string        `json:"target_commitish"`
		Draft      bool          `json:"draft"`
		Prerelease bool          `json:"prerelease"`
		Body       string        `json:"body"`
		URL        string        `json:"url"`
		HTMLURL    string        `json:"html_url"`
		AssetsURL  string        `json:"assets_url"`
		UploadURL  string        `json:"upload_url"`
		Assets     []GitHubAsset `json:"assets"`
	}

	// GitHubAsset represents an asset for a GitHub release.
	GitHubAsset struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Label       string `json:"label"`
		State       string `json:"state"`
		Size        int    `json:"size"`
		ContentType string `json:"content_type"`
		URL         string `json:"url"`
		DownloadURL string `json:"browser_download_url"`
	}
)

// httpError is an http error.
type httpError struct {
	Request    *http.Request
	StatusCode int
	Message    string
}

// newHTTPError creates a new instance of httpError.
func newHTTPError(res *http.Response) *httpError {
	err := &httpError{
		Request:    res.Request,
		StatusCode: res.StatusCode,
	}

	if res.Body != nil {
		if data, e := ioutil.ReadAll(res.Body); e == nil {
			err.Message = string(data)
		}
	}

	return err
}

func (e *httpError) Error() string {
	return fmt.Sprintf("%s %s %d: %s", e.Request.Method, e.Request.URL.Path, e.StatusCode, e.Message)
}

// GitHubBranchProtection enables/disables branch protection for administrators.
// See https://developer.github.com/v3/repos/branches/#get-admin-enforcement-of-protected-branch
// See https://developer.github.com/v3/repos/branches/#add-admin-enforcement-of-protected-branch
// See https://developer.github.com/v3/repos/branches/#remove-admin-enforcement-of-protected-branch
type GitHubBranchProtection struct {
	Mock    Step
	Client  *http.Client
	Token   string
	BaseURL string
	Repo    string
	Branch  string
	Enabled bool
}

func (s *GitHubBranchProtection) makeRequest(ctx context.Context, method string) (map[string]interface{}, error) {
	var statusCode int
	switch method {
	case "GET", "POST":
		statusCode = 200
	case "DELETE":
		statusCode = 204
	}

	url := fmt.Sprintf("%s/repos/%s/branches/%s/protection/enforce_admins", s.BaseURL, s.Repo, s.Branch)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	req.Header.Set("Authorization", fmt.Sprintf("token %s", s.Token))
	req.Header.Set("Accept", githubAcceptType)
	req.Header.Set("User-Agent", githubUserAgent) // ref: https://developer.github.com/v3/#user-agent-required

	res, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != statusCode {
		return nil, newHTTPError(res)
	}

	if method == "DELETE" {
		return nil, nil
	}

	var body map[string]interface{}
	body = map[string]interface{}{}
	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// Dry is a dry run of the step.
func (s *GitHubBranchProtection) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	_, err := s.makeRequest(ctx, "GET")
	if err != nil {
		return err
	}

	return nil
}

// Run executes the step.
func (s *GitHubBranchProtection) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	var method string
	if s.Enabled {
		method = "POST"
	} else {
		method = "DELETE"
	}

	_, err := s.makeRequest(ctx, method)
	if err != nil {
		return err
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitHubBranchProtection) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	var method string
	if s.Enabled {
		method = "DELETE"
	} else {
		method = "POST"
	}

	_, err := s.makeRequest(ctx, method)
	if err != nil {
		return err
	}

	return nil
}

// GitHubGetLatestRelease gets the latest release.
// See https://developer.github.com/v3/repos/releases/#get-the-latest-release
type GitHubGetLatestRelease struct {
	Mock    Step
	Client  *http.Client
	Token   string
	BaseURL string
	Repo    string
	Result  struct {
		LatestRelease GitHubRelease
	}
}

func (s *GitHubGetLatestRelease) makeRequest(ctx context.Context) (*GitHubRelease, error) {
	method := "GET"
	url := fmt.Sprintf("%s/repos/%s/releases/latest", s.BaseURL, s.Repo)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	req.Header.Set("Authorization", fmt.Sprintf("token %s", s.Token))
	req.Header.Set("Accept", githubAcceptType)
	req.Header.Set("User-Agent", githubUserAgent) // ref: https://developer.github.com/v3/#user-agent-required
	req.Header.Set("Content-Type", "application/json")

	res, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, newHTTPError(res)
	}

	release := new(GitHubRelease)
	err = json.NewDecoder(res.Body).Decode(release)
	if err != nil {
		return nil, err
	}

	return release, nil
}

// Dry is a dry run of the step.
func (s *GitHubGetLatestRelease) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	_, err := s.makeRequest(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Run executes the step.
func (s *GitHubGetLatestRelease) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	release, err := s.makeRequest(ctx)
	if err != nil {
		return err
	}

	s.Result.LatestRelease = *release

	return nil
}

// Revert reverts back an executed step.
func (s *GitHubGetLatestRelease) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	return nil
}

// GitHubCreateRelease creates a new GitHub release.
// See https://developer.github.com/v3/repos/releases/#get-the-latest-release
// See https://developer.github.com/v3/repos/releases/#create-a-release
// See https://developer.github.com/v3/repos/releases/#delete-a-release
type GitHubCreateRelease struct {
	Mock        Step
	Client      *http.Client
	Token       string
	BaseURL     string
	Repo        string
	ReleaseData GitHubReleaseData
	Result      struct {
		Release GitHubRelease
	}
}

func (s *GitHubCreateRelease) makeRequest(ctx context.Context, method, url string, body io.Reader) (*GitHubRelease, error) {
	var statusCode int
	switch method {
	case "GET":
		statusCode = 200
	case "POST":
		statusCode = 201
	case "DELETE":
		statusCode = 204
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	req.Header.Set("Authorization", fmt.Sprintf("token %s", s.Token))
	req.Header.Set("Accept", githubAcceptType)
	req.Header.Set("User-Agent", githubUserAgent) // ref: https://developer.github.com/v3/#user-agent-required
	req.Header.Set("Content-Type", "application/json")

	res, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != statusCode {
		return nil, newHTTPError(res)
	}

	if method == "DELETE" {
		return nil, nil
	}

	release := new(GitHubRelease)
	err = json.NewDecoder(res.Body).Decode(release)
	if err != nil {
		return nil, err
	}

	return release, nil
}

// Dry is a dry run of the step.
func (s *GitHubCreateRelease) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	method := "GET"
	url := fmt.Sprintf("%s/repos/%s/releases/latest", s.BaseURL, s.Repo)

	_, err := s.makeRequest(ctx, method, url, nil)
	if err != nil {
		return err
	}

	return nil
}

// Run executes the step.
func (s *GitHubCreateRelease) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	method := "POST"
	url := fmt.Sprintf("%s/repos/%s/releases", s.BaseURL, s.Repo)

	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(s.ReleaseData)

	release, err := s.makeRequest(ctx, method, url, body)
	if err != nil {
		return err
	}

	s.Result.Release = *release

	return nil
}

// Revert reverts back an executed step.
func (s *GitHubCreateRelease) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	method := "DELETE"
	url := fmt.Sprintf("%s/repos/%s/releases/%d", s.BaseURL, s.Repo, s.Result.Release.ID)

	_, err := s.makeRequest(ctx, method, url, nil)
	if err != nil {
		return err
	}

	return nil
}

// GitHubEditRelease edits an existing GitHub release.
// See https://developer.github.com/v3/repos/releases/#get-a-single-release
// See https://developer.github.com/v3/repos/releases/#edit-a-release
type GitHubEditRelease struct {
	Mock        Step
	Client      *http.Client
	Token       string
	BaseURL     string
	Repo        string
	ReleaseID   int
	ReleaseData GitHubReleaseData
	Result      struct {
		Release GitHubRelease
	}
}

func (s *GitHubEditRelease) makeRequest(ctx context.Context, method, url string, body io.Reader) (*GitHubRelease, error) {
	var statusCode int
	switch method {
	case "GET":
		statusCode = 200
	case "PATCH":
		statusCode = 200
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	req.Header.Set("Authorization", fmt.Sprintf("token %s", s.Token))
	req.Header.Set("Accept", githubAcceptType)
	req.Header.Set("User-Agent", githubUserAgent) // ref: https://developer.github.com/v3/#user-agent-required
	req.Header.Set("Content-Type", "application/json")

	res, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != statusCode {
		return nil, newHTTPError(res)
	}

	release := new(GitHubRelease)
	err = json.NewDecoder(res.Body).Decode(release)
	if err != nil {
		return nil, err
	}

	return release, nil
}

// Dry is a dry run of the step.
func (s *GitHubEditRelease) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	method := "GET"
	url := fmt.Sprintf("%s/repos/%s/releases/%d", s.BaseURL, s.Repo, s.ReleaseID)

	_, err := s.makeRequest(ctx, method, url, nil)
	if err != nil {
		return err
	}

	return nil
}

// Run executes the step.
func (s *GitHubEditRelease) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	method := "PATCH"
	url := fmt.Sprintf("%s/repos/%s/releases/%d", s.BaseURL, s.Repo, s.ReleaseID)

	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(s.ReleaseData)

	release, err := s.makeRequest(ctx, method, url, body)
	if err != nil {
		return err
	}

	s.Result.Release = *release

	return nil
}

// Revert reverts back an executed step.
func (s *GitHubEditRelease) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	// TODO: how to revert an edited release?
	return nil
}

// GitHubUploadAssets uploads assets (files) to GitHub for a release.
// See https://developer.github.com/v3/repos/releases/#list-assets-for-a-release
// See https://developer.github.com/v3/repos/releases/#upload-a-release-asset
// See https://developer.github.com/v3/repos/releases/#delete-a-release-asset
type GitHubUploadAssets struct {
	Mock             Step
	Client           *http.Client
	Token            string
	BaseURL          string
	Repo             string
	ReleaseID        int
	ReleaseUploadURL string
	AssetFiles       []string
	Result           struct {
		Assets []GitHubAsset
	}
}

// Dry is a dry run of the step
func (s *GitHubUploadAssets) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	method := "GET"
	url := fmt.Sprintf("%s/repos/%s/releases/%d/assets", s.BaseURL, s.Repo, s.ReleaseID)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)

	req.Header.Set("Authorization", fmt.Sprintf("token %s", s.Token))
	req.Header.Set("Accept", githubAcceptType)
	req.Header.Set("User-Agent", githubUserAgent) // ref: https://developer.github.com/v3/#user-agent-required
	req.Header.Set("Content-Type", "application/json")

	res, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return newHTTPError(res)
	}

	return nil
}

// Run executes the step
func (s *GitHubUploadAssets) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	s.Result.Assets = make([]GitHubAsset, 0)

	doneCh := make(chan error, len(s.AssetFiles))

	for _, file := range s.AssetFiles {
		go func(file string) {
			assetPath := filepath.Clean(file)
			assetName := filepath.Base(assetPath)

			method := "POST"
			re := regexp.MustCompile(`\{\?[0-9A-Za-z_,]+\}`)
			url := re.ReplaceAllLiteralString(s.ReleaseUploadURL, "")
			url = fmt.Sprintf("%s?name=%s", url, netURL.QueryEscape(assetName))

			content, err := getUploadContent(assetPath)
			if err != nil {
				doneCh <- err
				return
			}
			defer content.Body.Close()

			req, err := http.NewRequest(method, url, content.Body)
			if err != nil {
				doneCh <- err
				return
			}

			req = req.WithContext(ctx)

			req.Header.Set("Authorization", fmt.Sprintf("token %s", s.Token))
			req.Header.Set("Accept", githubAcceptType)
			req.Header.Set("User-Agent", githubUserAgent) // ref: https://developer.github.com/v3/#user-agent-required
			req.Header.Set("Content-Type", content.MIMEType)
			req.ContentLength = content.Length

			res, err := s.Client.Do(req)
			if err != nil {
				doneCh <- err
				return
			}
			defer res.Body.Close()

			if res.StatusCode != 201 {
				doneCh <- newHTTPError(res)
				return
			}

			asset := GitHubAsset{}
			err = json.NewDecoder(res.Body).Decode(&asset)
			if err != nil {
				doneCh <- err
				return
			}

			s.Result.Assets = append(s.Result.Assets, asset)
			doneCh <- nil
		}(file)
	}

	for range s.AssetFiles {
		if err := <-doneCh; err != nil {
			return err
		}
	}

	return nil
}

// Revert reverts back an executed step
func (s *GitHubUploadAssets) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	for _, asset := range s.Result.Assets {
		method := "DELETE"
		url := fmt.Sprintf("%s/repos/%s/releases/assets/%d", s.BaseURL, s.Repo, asset.ID)

		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return err
		}

		req = req.WithContext(ctx)

		req.Header.Set("Authorization", fmt.Sprintf("token %s", s.Token))
		req.Header.Set("Accept", githubAcceptType)
		req.Header.Set("User-Agent", githubUserAgent) // ref: https://developer.github.com/v3/#user-agent-required
		req.Header.Set("Content-Type", "application/json")

		res, err := s.Client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != 204 {
			return newHTTPError(res)
		}
	}

	return nil
}

type uploadContent struct {
	Body     io.ReadCloser
	Length   int64
	MIMEType string
}

func getUploadContent(filepath string) (*uploadContent, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Read the first 512 bytes of file to determine the mime type of asset
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		return nil, err
	}

	// Determine mime type of asset
	// http.DetectContentType will return "application/octet-stream" if it cannot determine a more specific one
	mimeType := http.DetectContentType(buff)

	// Reset the offset back to the beginning of the file
	// SEEK_SET: seek relative to the origin of the file
	_, err = file.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, err
	}

	return &uploadContent{
		Body:     file,
		Length:   stat.Size(),
		MIMEType: mimeType,
	}, nil
}

// GitHubDownloadAsset downloads an asset file and writes to a local file.
type GitHubDownloadAsset struct {
	Mock      Step
	Client    *http.Client
	Token     string
	BaseURL   string
	Repo      string
	Tag       string
	AssetName string
	Filepath  string
	Result    struct {
		Size int64
	}
}

func (s *GitHubDownloadAsset) makeRequest(ctx context.Context) (io.ReadCloser, error) {
	method := "GET"
	url := fmt.Sprintf("%s/%s/releases/download/%s/%s", s.BaseURL, s.Repo, s.Tag, s.AssetName)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	req.Header.Set("Authorization", fmt.Sprintf("token %s", s.Token))
	req.Header.Set("User-Agent", githubUserAgent) // ref: https://developer.github.com/v3/#user-agent-required

	res, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, newHTTPError(res)
	}

	return res.Body, nil
}

// Dry is a dry run of the step.
func (s *GitHubDownloadAsset) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	body, err := s.makeRequest(ctx)
	if err != nil {
		return err
	}
	defer body.Close()

	return nil
}

// Run executes the step.
func (s *GitHubDownloadAsset) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	body, err := s.makeRequest(ctx)
	if err != nil {
		return err
	}
	defer body.Close()

	file, err := os.OpenFile(s.Filepath, os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	size, err := io.Copy(file, body)
	if err != nil {
		return err
	}

	s.Result.Size = size

	return nil
}

// Revert reverts back an executed step.
func (s *GitHubDownloadAsset) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	return os.Remove(s.Filepath)
}