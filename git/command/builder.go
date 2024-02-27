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

package command

import (
	"fmt"
	"strings"
)

const (
	// NoRefUpdates denotes a command which will never update refs.
	NoRefUpdates = 1 << iota
	// NoEndOfOptions denotes a command which doesn't know --end-of-options.
	NoEndOfOptions
)

type builder struct {
	flags                  uint
	validatePositionalArgs func([]string) error
}

// supportsEndOfOptions indicates whether a command can handle the
// `--end-of-options` option.
func (b builder) supportsEndOfOptions() bool {
	return b.flags&NoEndOfOptions == 0
}

// descriptions is a curated list of Git command descriptions.
var descriptions = map[string]builder{
	"am": {},
	"apply": {
		flags: NoRefUpdates,
	},
	"archive": {
		// git-archive(1) does not support disambiguating options from paths from revisions.
		flags: NoRefUpdates | NoEndOfOptions,
	},
	"blame": {
		// git-blame(1) does not support disambiguating options from paths from revisions.
		flags: NoRefUpdates | NoEndOfOptions,
	},
	"bundle": {
		flags: NoRefUpdates,
	},
	"cat-file": {
		flags: NoRefUpdates,
	},
	"check-attr": {
		flags: NoRefUpdates | NoEndOfOptions,
	},
	"check-ref-format": {
		// git-check-ref-format(1) uses a hand-rolled option parser which doesn't support
		// `--end-of-options`.
		flags: NoRefUpdates | NoEndOfOptions,
	},
	"checkout": {
		// git-checkout(1) does not support disambiguating options from paths from
		// revisions.
		flags: NoEndOfOptions,
	},
	"clone": {},
	"commit": {
		flags: 0,
	},
	"commit-graph": {
		flags: NoRefUpdates,
	},
	"commit-tree": {
		flags: NoRefUpdates,
	},
	"config": {
		flags: NoRefUpdates,
	},
	"count-objects": {
		flags: NoRefUpdates,
	},
	"diff": {
		flags: NoRefUpdates,
	},
	"diff-tree": {
		flags: NoRefUpdates,
	},
	"fetch": {
		flags: 0,
	},
	"for-each-ref": {
		flags: NoRefUpdates,
	},
	"format-patch": {
		flags: NoRefUpdates,
	},
	"fsck": {
		flags: NoRefUpdates,
	},
	"gc": {
		flags: NoRefUpdates,
	},
	"grep": {
		// git-grep(1) does not support disambiguating options from paths from
		// revisions.
		flags: NoRefUpdates | NoEndOfOptions,
	},
	"hash-object": {
		flags: NoRefUpdates,
	},
	"index-pack": {
		flags: NoRefUpdates | NoEndOfOptions,
	},
	"init": {
		flags: NoRefUpdates,
	},
	"log": {
		flags: NoRefUpdates,
	},
	"ls-remote": {
		flags: NoRefUpdates,
	},
	"ls-tree": {
		flags: NoRefUpdates,
	},
	"merge-base": {
		flags: NoRefUpdates,
	},
	"merge-file": {
		flags: NoRefUpdates,
	},
	"merge-tree": {
		flags: NoRefUpdates,
	},
	"mktag": {
		flags: NoRefUpdates,
	},
	"mktree": {
		flags: NoRefUpdates,
	},
	"multi-pack-index": {
		flags: NoRefUpdates,
	},
	"pack-refs": {
		flags: NoRefUpdates,
	},
	"pack-objects": {
		flags: NoRefUpdates,
	},
	"patch-id": {
		flags: NoRefUpdates | NoEndOfOptions,
	},
	"prune": {
		flags: NoRefUpdates,
	},
	"prune-packed": {
		flags: NoRefUpdates,
	},
	"push": {
		flags: NoRefUpdates,
	},
	"read-tree": {
		flags: NoRefUpdates,
	},
	"receive-pack": {
		flags: 0,
	},
	"remote": {
		// While git-remote(1)'s `add` subcommand does support `--end-of-options`,
		// `remove` doesn't.
		flags: NoEndOfOptions,
	},
	"repack": {
		flags: NoRefUpdates,
	},
	"rev-list": {
		// We cannot use --end-of-options here because pseudo revisions like `--all`
		// and `--not` count as options.
		flags: NoRefUpdates | NoEndOfOptions,
		validatePositionalArgs: func(args []string) error {
			for _, arg := range args {
				// git-rev-list(1) supports pseudo-revision arguments which can be
				// intermingled with normal positional arguments. Given that these
				// pseudo-revisions have leading dashes, normal validation would
				// refuse them as positional arguments. We thus override validation
				// for two of these which we are using in our codebase. There are
				// more, but we can add them at a later point if they're ever
				// required.
				if arg == "--all" || arg == "--not" {
					continue
				}
				if err := validatePositionalArg(arg); err != nil {
					return fmt.Errorf("rev-list: %w", err)
				}
			}
			return nil
		},
	},
	"rev-parse": {
		// --end-of-options is echoed by git-rev-parse(1) if used without
		// `--verify`.
		flags: NoRefUpdates | NoEndOfOptions,
	},
	"show": {
		flags: NoRefUpdates,
	},
	"show-ref": {
		flags: NoRefUpdates,
	},
	"symbolic-ref": {
		flags: 0,
	},
	"tag": {
		flags: 0,
	},
	"unpack-objects": {
		flags: NoRefUpdates | NoEndOfOptions,
	},
	"update-ref": {
		flags: 0,
	},
	"upload-archive": {
		// git-upload-archive(1) has a handrolled parser which always interprets the
		// first argument as directory, so we cannot use `--end-of-options`.
		flags: NoRefUpdates | NoEndOfOptions,
	},
	"upload-pack": {
		flags: NoRefUpdates,
	},
	"version": {
		flags: NoRefUpdates,
	},
	"worktree": {
		flags: 0,
	},
}

// args validates the given flags and arguments and, if valid, returns the complete command line.
func (b builder) args(flags []string, args []string, postSepArgs []string) ([]string, error) {
	var cmdArgs []string

	cmdArgs = append(cmdArgs, flags...)

	if b.supportsEndOfOptions() {
		cmdArgs = append(cmdArgs, "--end-of-options")
	}

	if b.validatePositionalArgs != nil {
		if err := b.validatePositionalArgs(args); err != nil {
			return nil, err
		}
	} else {
		for _, a := range args {
			if err := validatePositionalArg(a); err != nil {
				return nil, err
			}
		}
	}
	cmdArgs = append(cmdArgs, args...)

	if len(postSepArgs) > 0 {
		cmdArgs = append(cmdArgs, "--")
	}

	// post separator args do not need any validation
	cmdArgs = append(cmdArgs, postSepArgs...)

	return cmdArgs, nil
}

func validatePositionalArg(arg string) error {
	if strings.HasPrefix(arg, "-") {
		return fmt.Errorf("positional arg %q cannot start with dash '-': %w", arg, ErrInvalidArg)
	}
	return nil
}
