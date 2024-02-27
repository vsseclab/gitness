// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package url

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
)

const (
	// GITSuffix is the suffix used to terminate repo paths for git apis.
	GITSuffix = ".git"

	// APIMount is the prefix path for the api endpoints.
	APIMount = "api"

	// GITMount is the prefix path for the git endpoints.
	GITMount = "git"
)

// Provider is an abstraction of a component that provides system related URLs.
// NOTE: Abstract to allow for custom implementation for more complex routing environments.
type Provider interface {
	// GetInternalAPIURL returns the internally reachable base url of the server.
	// NOTE: url is guaranteed to not have any trailing '/'.
	GetInternalAPIURL() string

	// GenerateContainerGITCloneURL generates a URL that can be used by CI container builds to
	// interact with gitness and clone a repo.
	GenerateContainerGITCloneURL(repoPath string) string

	// GenerateGITCloneURL generates the public git clone URL for the provided repo path.
	// NOTE: url is guaranteed to not have any trailing '/'.
	GenerateGITCloneURL(repoPath string) string

	// GenerateUIRepoURL returns the url for the UI screen of a repository.
	GenerateUIRepoURL(repoPath string) string

	// GenerateUIPRURL returns the url for the UI screen of an existing pr.
	GenerateUIPRURL(repoPath string, prID int64) string

	// GenerateUICompareURL returns the url for the UI screen comparing two references.
	GenerateUICompareURL(repoPath string, ref1 string, ref2 string) string

	// GetAPIHostname returns the host for the api endpoint.
	GetAPIHostname() string

	// GenerateUIBuildURL returns the endpoint to use for viewing build executions.
	GenerateUIBuildURL(repoPath, pipelineIdentifier string, seqNumber int64) string

	// GetGITHostname returns the host for the git endpoint.
	GetGITHostname() string

	// GetAPIProto returns the proto for the API hostname
	GetAPIProto() string
}

// Provider provides the URLs of the gitness system.
type provider struct {
	// internalURL stores the URL via which the service is reachable at internally
	// (no need for internal services to go via public route).
	internalURL *url.URL

	// containerURL stores the URL that can be used to communicate with gitness from inside a
	// build container.
	containerURL *url.URL

	// apiURL stores the raw URL the api endpoints are reachable at publicly.
	apiURL *url.URL

	// gitURL stores the URL the git endpoints are available at.
	// NOTE: we store it as url.URL so we can derive clone URLS without errors.
	gitURL *url.URL

	// uiURL stores the raw URL to the ui endpoints.
	uiURL *url.URL
}

func NewProvider(
	internalURLRaw,
	containerURLRaw string,
	apiURLRaw string,
	gitURLRaw,
	uiURLRaw string,
) (Provider, error) {
	// remove trailing '/' to make usage easier
	internalURLRaw = strings.TrimRight(internalURLRaw, "/")
	containerURLRaw = strings.TrimRight(containerURLRaw, "/")
	apiURLRaw = strings.TrimRight(apiURLRaw, "/")
	gitURLRaw = strings.TrimRight(gitURLRaw, "/")
	uiURLRaw = strings.TrimRight(uiURLRaw, "/")

	internalURL, err := url.Parse(internalURLRaw)
	if err != nil {
		return nil, fmt.Errorf("provided internalURLRaw '%s' is invalid: %w", internalURLRaw, err)
	}

	containerURL, err := url.Parse(containerURLRaw)
	if err != nil {
		return nil, fmt.Errorf("provided containerURLRaw '%s' is invalid: %w", containerURLRaw, err)
	}

	apiURL, err := url.Parse(apiURLRaw)
	if err != nil {
		return nil, fmt.Errorf("provided apiURLRaw '%s' is invalid: %w", apiURLRaw, err)
	}

	gitURL, err := url.Parse(gitURLRaw)
	if err != nil {
		return nil, fmt.Errorf("provided gitURLRaw '%s' is invalid: %w", gitURLRaw, err)
	}

	uiURL, err := url.Parse(uiURLRaw)
	if err != nil {
		return nil, fmt.Errorf("provided uiURLRaw '%s' is invalid: %w", uiURLRaw, err)
	}

	return &provider{
		internalURL:  internalURL,
		containerURL: containerURL,
		apiURL:       apiURL,
		gitURL:       gitURL,
		uiURL:        uiURL,
	}, nil
}

func (p *provider) GetInternalAPIURL() string {
	return p.internalURL.JoinPath(APIMount).String()
}

func (p *provider) GenerateContainerGITCloneURL(repoPath string) string {
	repoPath = path.Clean(repoPath)
	if !strings.HasSuffix(repoPath, GITSuffix) {
		repoPath += GITSuffix
	}

	return p.containerURL.JoinPath(GITMount, repoPath).String()
}

func (p *provider) GenerateGITCloneURL(repoPath string) string {
	repoPath = path.Clean(repoPath)
	if !strings.HasSuffix(repoPath, GITSuffix) {
		repoPath += GITSuffix
	}

	return p.gitURL.JoinPath(repoPath).String()
}

func (p *provider) GenerateUIBuildURL(repoPath, pipelineIdentifier string, seqNumber int64) string {
	return p.uiURL.JoinPath(repoPath, "pipelines",
		pipelineIdentifier, "execution", strconv.Itoa(int(seqNumber))).String()
}

func (p *provider) GenerateUIRepoURL(repoPath string) string {
	return p.uiURL.JoinPath(repoPath).String()
}

func (p *provider) GenerateUIPRURL(repoPath string, prID int64) string {
	return p.uiURL.JoinPath(repoPath, "pulls", fmt.Sprint(prID)).String()
}

func (p *provider) GenerateUICompareURL(repoPath string, ref1 string, ref2 string) string {
	return p.uiURL.JoinPath(repoPath, "pulls/compare", ref1+"..."+ref2).String()
}

func (p *provider) GetAPIHostname() string {
	return p.apiURL.Hostname()
}

func (p *provider) GetGITHostname() string {
	return p.gitURL.Hostname()
}

func (p *provider) GetAPIProto() string {
	return p.apiURL.Scheme
}
