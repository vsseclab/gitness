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

package git

import (
	"os"
	"path/filepath"

	"github.com/harness/gitness/errors"
	"github.com/harness/gitness/git/storage"
	"github.com/harness/gitness/git/types"
)

const (
	repoSubdirName           = "repos"
	ReposGraveyardSubdirName = "cleanup"
)

type Service struct {
	reposRoot      string
	tmpDir         string
	adapter        Adapter
	store          storage.Store
	gitHookPath    string
	reposGraveyard string
}

func New(
	config types.Config,
	adapter Adapter,
	storage storage.Store,
) (*Service, error) {
	// Create repos folder
	reposRoot := filepath.Join(config.Root, repoSubdirName)
	if _, err := os.Stat(reposRoot); errors.Is(err, os.ErrNotExist) {
		if err = os.MkdirAll(reposRoot, fileMode700); err != nil {
			return nil, err
		}
	}

	// create a temp dir for deleted repositories
	// this dir should get cleaned up peridocally if it's not empty
	reposGraveyard := filepath.Join(config.Root, ReposGraveyardSubdirName)
	if _, errdir := os.Stat(reposGraveyard); os.IsNotExist(errdir) {
		if errdir = os.MkdirAll(reposGraveyard, fileMode700); errdir != nil {
			return nil, errdir
		}
	}
	return &Service{
		reposRoot:      reposRoot,
		tmpDir:         config.TmpDir,
		reposGraveyard: reposGraveyard,
		adapter:        adapter,
		store:          storage,
		gitHookPath:    config.HookPath,
	}, nil
}
