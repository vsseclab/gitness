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

package reposize

import (
	"github.com/harness/gitness/app/store"
	"github.com/harness/gitness/git"
	"github.com/harness/gitness/job"
	"github.com/harness/gitness/types"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideCalculator,
)

func ProvideCalculator(
	config *types.Config,
	git git.Interface,
	repoStore store.RepoStore,
	scheduler *job.Scheduler,
	executor *job.Executor,
) (*Calculator, error) {
	job := &Calculator{
		enabled:    config.RepoSize.Enabled,
		cron:       config.RepoSize.CRON,
		maxDur:     config.RepoSize.MaxDuration,
		numWorkers: config.RepoSize.NumWorkers,
		git:        git,
		repoStore:  repoStore,
		scheduler:  scheduler,
	}

	err := executor.Register(jobType, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}
