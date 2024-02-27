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

package pullreq

import (
	"context"
	"fmt"
	"strconv"

	pullreqevents "github.com/harness/gitness/app/events/pullreq"
	"github.com/harness/gitness/events"
	"github.com/harness/gitness/git"
	gitenum "github.com/harness/gitness/git/enum"
)

// createHeadRefOnCreated handles pull request Created events.
// It creates the PR head git ref.
func (s *Service) createHeadRefOnCreated(ctx context.Context,
	event *events.Event[*pullreqevents.CreatedPayload],
) error {
	repoGit, err := s.repoGitInfoCache.Get(ctx, event.Payload.TargetRepoID)
	if err != nil {
		return fmt.Errorf("failed to get repo git info: %w", err)
	}

	writeParams, err := createSystemRPCWriteParams(ctx, s.urlProvider, repoGit.ID, repoGit.GitUID)
	if err != nil {
		return fmt.Errorf("failed to generate rpc write params: %w", err)
	}

	// TODO: This doesn't work for forked repos (only works when sourceRepo==targetRepo).
	// This is because commits from the source repository must be first pulled into the target repository.
	err = s.git.UpdateRef(ctx, git.UpdateRefParams{
		WriteParams: writeParams,
		Name:        strconv.Itoa(int(event.Payload.Number)),
		Type:        gitenum.RefTypePullReqHead,
		NewValue:    event.Payload.SourceSHA,
		OldValue:    "", // this is a new pull request, so we expect that the ref doesn't exist
	})
	if err != nil {
		return fmt.Errorf("failed to update PR head ref: %w", err)
	}

	return nil
}

// updateHeadRefOnBranchUpdate handles pull request Branch Updated events.
// It updates the PR head git ref to point to the latest commit.
func (s *Service) updateHeadRefOnBranchUpdate(ctx context.Context,
	event *events.Event[*pullreqevents.BranchUpdatedPayload],
) error {
	repoGit, err := s.repoGitInfoCache.Get(ctx, event.Payload.TargetRepoID)
	if err != nil {
		return fmt.Errorf("failed to get repo git info: %w", err)
	}

	writeParams, err := createSystemRPCWriteParams(ctx, s.urlProvider, repoGit.ID, repoGit.GitUID)
	if err != nil {
		return fmt.Errorf("failed to generate rpc write params: %w", err)
	}

	// TODO: This doesn't work for forked repos (only works when sourceRepo==targetRepo)
	// This is because commits from the source repository must be first pulled into the target repository.
	err = s.git.UpdateRef(ctx, git.UpdateRefParams{
		WriteParams: writeParams,
		Name:        strconv.Itoa(int(event.Payload.Number)),
		Type:        gitenum.RefTypePullReqHead,
		NewValue:    event.Payload.NewSHA,
		OldValue:    event.Payload.OldSHA,
	})
	if err != nil {
		return fmt.Errorf("failed to update PR head ref after new commit: %w", err)
	}

	return nil
}

// updateHeadRefOnReopen handles pull request StateChanged events.
// It updates the PR head git ref to point to the source branch commit SHA.
func (s *Service) updateHeadRefOnReopen(ctx context.Context,
	event *events.Event[*pullreqevents.ReopenedPayload],
) error {
	repoGit, err := s.repoGitInfoCache.Get(ctx, event.Payload.TargetRepoID)
	if err != nil {
		return fmt.Errorf("failed to get repo git info: %w", err)
	}

	writeParams, err := createSystemRPCWriteParams(ctx, s.urlProvider, repoGit.ID, repoGit.GitUID)
	if err != nil {
		return fmt.Errorf("failed to generate rpc write params: %w", err)
	}

	// TODO: This doesn't work for forked repos (only works when sourceRepo==targetRepo)
	// This is because commits from the source repository must be first pulled into the target repository.
	err = s.git.UpdateRef(ctx, git.UpdateRefParams{
		WriteParams: writeParams,
		Name:        strconv.Itoa(int(event.Payload.Number)),
		Type:        gitenum.RefTypePullReqHead,
		NewValue:    event.Payload.SourceSHA,
		OldValue:    "", // the request is re-opened, so anything can be the old value
	})
	if err != nil {
		return fmt.Errorf("failed to update PR head ref after pull request reopen: %w", err)
	}

	return nil
}
