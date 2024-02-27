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

package authz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/harness/gitness/app/paths"
	"github.com/harness/gitness/app/store"
	"github.com/harness/gitness/cache"
	gitness_store "github.com/harness/gitness/store"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/enum"

	"golang.org/x/exp/slices"
)

type PermissionCacheKey struct {
	PrincipalID int64
	SpaceRef    string
	Permission  enum.Permission
}
type PermissionCache cache.Cache[PermissionCacheKey, bool]

func NewPermissionCache(
	spaceStore store.SpaceStore,
	membershipStore store.MembershipStore,
	cacheDuration time.Duration,
) PermissionCache {
	return cache.New[PermissionCacheKey, bool](permissionCacheGetter{
		spaceStore:      spaceStore,
		membershipStore: membershipStore,
	}, cacheDuration)
}

type permissionCacheGetter struct {
	spaceStore      store.SpaceStore
	membershipStore store.MembershipStore
}

func (g permissionCacheGetter) Find(ctx context.Context, key PermissionCacheKey) (bool, error) {
	spaceRef := key.SpaceRef
	principalID := key.PrincipalID

	// Find the starting space.
	space, err := g.spaceStore.FindByRef(ctx, spaceRef)
	if err != nil {
		return false, fmt.Errorf("failed to find space '%s': %w", spaceRef, err)
	}

	// limit the depth to be safe (e.g. root/space1/space2 => maxDepth of 3)
	maxDepth := len(paths.Segments(spaceRef))

	for depth := 0; depth < maxDepth; depth++ {
		// Find the membership in the current space.
		membership, err := g.membershipStore.Find(ctx, types.MembershipKey{
			SpaceID:     space.ID,
			PrincipalID: principalID,
		})
		if err != nil && !errors.Is(err, gitness_store.ErrResourceNotFound) {
			return false, fmt.Errorf("failed to find membership: %w", err)
		}

		// If the membership is defined in the current space, check if the user has the required permission.
		if membership != nil &&
			roleHasPermission(membership.Role, key.Permission) {
			return true, nil
		}

		// If membership with the requested permission has not been found in the current space,
		// move to the parent space, if any.

		if space.ParentID == 0 {
			return false, nil
		}

		space, err = g.spaceStore.Find(ctx, space.ParentID)
		if err != nil {
			return false, fmt.Errorf("failed to find parent space with id %d: %w", space.ParentID, err)
		}
	}

	return false, nil
}

func roleHasPermission(role enum.MembershipRole, permission enum.Permission) bool {
	_, hasRole := slices.BinarySearch(role.Permissions(), permission)
	return hasRole
}
