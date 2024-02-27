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

package repo

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	apiauth "github.com/harness/gitness/app/api/auth"
	"github.com/harness/gitness/app/api/controller/limiter"
	"github.com/harness/gitness/app/api/usererror"
	"github.com/harness/gitness/app/auth"
	"github.com/harness/gitness/app/bootstrap"
	"github.com/harness/gitness/app/githook"
	"github.com/harness/gitness/git"
	"github.com/harness/gitness/resources"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/check"
	"github.com/harness/gitness/types/enum"

	"github.com/rs/zerolog/log"
)

var (
	// errRepositoryRequiresParent if the user tries to create a repo without a parent space.
	errRepositoryRequiresParent = usererror.BadRequest(
		"Parent space required - standalone repositories are not supported.")
)

type CreateInput struct {
	ParentRef string `json:"parent_ref"`
	// TODO [CODE-1363]: remove after identifier migration.
	UID           string `json:"uid" deprecated:"true"`
	Identifier    string `json:"identifier"`
	DefaultBranch string `json:"default_branch"`
	Description   string `json:"description"`
	IsPublic      bool   `json:"is_public"`
	ForkID        int64  `json:"fork_id"`
	Readme        bool   `json:"readme"`
	License       string `json:"license"`
	GitIgnore     string `json:"git_ignore"`
}

// Create creates a new repository.
//
//nolint:gocognit
func (c *Controller) Create(ctx context.Context, session *auth.Session, in *CreateInput) (*types.Repository, error) {
	if err := c.sanitizeCreateInput(in); err != nil {
		return nil, fmt.Errorf("failed to sanitize input: %w", err)
	}

	parentSpace, err := c.getSpaceCheckAuthRepoCreation(ctx, session, in.ParentRef)
	if err != nil {
		return nil, err
	}

	var repo *types.Repository
	err = c.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := c.resourceLimiter.RepoCount(ctx, parentSpace.ID, 1); err != nil {
			return fmt.Errorf("resource limit exceeded: %w", limiter.ErrMaxNumReposReached)
		}

		gitResp, err := c.createGitRepository(ctx, session, in)
		if err != nil {
			return fmt.Errorf("error creating repository on git: %w", err)
		}

		now := time.Now().UnixMilli()
		repo = &types.Repository{
			Version:       0,
			ParentID:      parentSpace.ID,
			Identifier:    in.Identifier,
			GitUID:        gitResp.UID,
			Description:   in.Description,
			IsPublic:      in.IsPublic,
			CreatedBy:     session.Principal.ID,
			Created:       now,
			Updated:       now,
			ForkID:        in.ForkID,
			DefaultBranch: in.DefaultBranch,
		}
		err = c.repoStore.Create(ctx, repo)
		if err != nil {
			if dErr := c.deleteGitRepository(ctx, session, repo); dErr != nil {
				log.Ctx(ctx).Warn().Err(dErr).Msg("failed to delete repo for cleanup")
			}
			return fmt.Errorf("failed to create repository in storage: %w", err)
		}

		return nil
	}, sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}

	// backfil GitURL
	repo.GitURL = c.urlProvider.GenerateGITCloneURL(repo.Path)

	// index repository if files are created
	if in.Readme || in.GitIgnore != "" || (in.License != "" && in.License != "none") {
		err = c.indexer.Index(ctx, repo)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Int64("repo_id", repo.ID).Msg("failed to index repo")
		}
	}

	return repo, nil
}

func (c *Controller) getSpaceCheckAuthRepoCreation(
	ctx context.Context,
	session *auth.Session,
	parentRef string,
) (*types.Space, error) {
	space, err := c.spaceStore.FindByRef(ctx, parentRef)
	if err != nil {
		return nil, fmt.Errorf("parent space not found: %w", err)
	}

	// create is a special case - check permission without specific resource
	scope := &types.Scope{SpacePath: space.Path}
	resource := &types.Resource{
		Type:       enum.ResourceTypeRepo,
		Identifier: "",
	}

	err = apiauth.Check(ctx, c.authorizer, session, scope, resource, enum.PermissionRepoEdit)
	if err != nil {
		return nil, fmt.Errorf("auth check failed: %w", err)
	}

	return space, nil
}

func (c *Controller) sanitizeCreateInput(in *CreateInput) error {
	// TODO [CODE-1363]: remove after identifier migration.
	if in.Identifier == "" {
		in.Identifier = in.UID
	}

	if in.IsPublic && !c.publicResourceCreationEnabled {
		return errPublicRepoCreationDisabled
	}

	if err := c.validateParentRef(in.ParentRef); err != nil {
		return err
	}

	if err := check.RepoIdentifier(in.Identifier); err != nil {
		return err
	}

	in.Description = strings.TrimSpace(in.Description)
	if err := check.Description(in.Description); err != nil {
		return err
	}

	if in.DefaultBranch == "" {
		in.DefaultBranch = c.defaultBranch
	}

	return nil
}

func (c *Controller) createGitRepository(ctx context.Context, session *auth.Session,
	in *CreateInput) (*git.CreateRepositoryOutput, error) {
	var (
		err     error
		content []byte
	)
	files := make([]git.File, 0, 3) // readme, gitignore, licence
	if in.Readme {
		content = createReadme(in.Identifier, in.Description)
		files = append(files, git.File{
			Path:    "README.md",
			Content: content,
		})
	}
	if in.License != "" && in.License != "none" {
		content, err = resources.ReadLicense(in.License)
		if err != nil {
			return nil, fmt.Errorf("failed to read license '%s': %w", in.License, err)
		}
		files = append(files, git.File{
			Path:    "LICENSE",
			Content: content,
		})
	}
	if in.GitIgnore != "" {
		content, err = resources.ReadGitIgnore(in.GitIgnore)
		if err != nil {
			return nil, fmt.Errorf("failed to read git ignore '%s': %w", in.GitIgnore, err)
		}
		files = append(files, git.File{
			Path:    ".gitignore",
			Content: content,
		})
	}

	// generate envars (add everything githook CLI needs for execution)
	envVars, err := githook.GenerateEnvironmentVariables(
		ctx,
		c.urlProvider.GetInternalAPIURL(),
		0,
		session.Principal.ID,
		true,
		true,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate git hook environment variables: %w", err)
	}

	actor := identityFromPrincipal(session.Principal)
	committer := identityFromPrincipal(bootstrap.NewSystemServiceSession().Principal)
	now := time.Now()
	resp, err := c.git.CreateRepository(ctx, &git.CreateRepositoryParams{
		Actor:         *actor,
		EnvVars:       envVars,
		DefaultBranch: in.DefaultBranch,
		Files:         files,
		Author:        actor,
		AuthorDate:    &now,
		Committer:     committer,
		CommitterDate: &now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create repo on: %w", err)
	}

	return resp, nil
}

func createReadme(name, description string) []byte {
	content := bytes.Buffer{}
	content.WriteString("# " + name + "\n")
	if description != "" {
		content.WriteString(description)
	}
	return content.Bytes()
}

func identityFromPrincipal(p types.Principal) *git.Identity {
	return &git.Identity{
		Name:  p.DisplayName,
		Email: p.Email,
	}
}
