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

package store

import (
	"github.com/harness/gitness/cache"
	"github.com/harness/gitness/types"
)

type (
	// PrincipalInfoCache caches principal IDs to principal info.
	PrincipalInfoCache cache.ExtendedCache[int64, *types.PrincipalInfo]

	// SpacePathCache caches a raw path to a space path.
	SpacePathCache cache.Cache[string, *types.SpacePath]

	// RepoGitInfoCache caches repository IDs to values GitUID.
	RepoGitInfoCache cache.Cache[int64, *types.RepositoryGitInfo]
)
