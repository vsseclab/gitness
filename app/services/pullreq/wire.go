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

	gitevents "github.com/harness/gitness/app/events/git"
	pullreqevents "github.com/harness/gitness/app/events/pullreq"
	"github.com/harness/gitness/app/services/codecomments"
	"github.com/harness/gitness/app/sse"
	"github.com/harness/gitness/app/store"
	"github.com/harness/gitness/app/url"
	"github.com/harness/gitness/events"
	"github.com/harness/gitness/git"
	"github.com/harness/gitness/pubsub"
	"github.com/harness/gitness/types"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideService,
)

func ProvideService(ctx context.Context,
	config *types.Config,
	gitReaderFactory *events.ReaderFactory[*gitevents.Reader],
	pullReqEvFactory *events.ReaderFactory[*pullreqevents.Reader],
	pullReqEvReporter *pullreqevents.Reporter,
	git git.Interface,
	repoGitInfoCache store.RepoGitInfoCache,
	repoStore store.RepoStore,
	pullreqStore store.PullReqStore,
	activityStore store.PullReqActivityStore,
	codeCommentView store.CodeCommentView,
	codeCommentMigrator *codecomments.Migrator,
	fileViewStore store.PullReqFileViewStore,
	pubsub pubsub.PubSub,
	urlProvider url.Provider,
	sseStreamer sse.Streamer,
) (*Service, error) {
	return New(ctx, config, gitReaderFactory, pullReqEvFactory, pullReqEvReporter, git,
		repoGitInfoCache, repoStore, pullreqStore, activityStore,
		codeCommentView, codeCommentMigrator, fileViewStore, pubsub, urlProvider, sseStreamer)
}
