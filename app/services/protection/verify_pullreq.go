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

package protection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/harness/gitness/app/services/codeowners"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/enum"

	"golang.org/x/exp/slices"
)

type (
	MergeVerifier interface {
		MergeVerify(ctx context.Context, in MergeVerifyInput) (MergeVerifyOutput, []types.RuleViolations, error)
		RequiredChecks(ctx context.Context, in RequiredChecksInput) (RequiredChecksOutput, error)
	}

	MergeVerifyInput struct {
		Actor        *types.Principal
		AllowBypass  bool
		IsRepoOwner  bool
		TargetRepo   *types.Repository
		SourceRepo   *types.Repository
		PullReq      *types.PullReq
		Reviewers    []*types.PullReqReviewer
		Method       enum.MergeMethod
		CheckResults []types.CheckResult
		CodeOwners   *codeowners.Evaluation
	}

	MergeVerifyOutput struct {
		DeleteSourceBranch bool
		AllowedMethods     []enum.MergeMethod
	}

	RequiredChecksInput struct {
		Actor       *types.Principal
		IsRepoOwner bool
		Repo        *types.Repository
		PullReq     *types.PullReq
	}

	RequiredChecksOutput struct {
		RequiredIdentifiers   map[string]struct{}
		BypassableIdentifiers map[string]struct{}
	}
)

// ensures that the DefPullReq type implements Sanitizer and MergeVerifier interface.
var (
	_ Sanitizer     = (*DefPullReq)(nil)
	_ MergeVerifier = (*DefPullReq)(nil)
)

const (
	codePullReqApprovalReqMinCount              = "pullreq.approvals.require_minimum_count"
	codePullReqApprovalReqMinCountLatest        = "pullreq.approvals.require_minimum_count:latest_commit"
	codePullReqApprovalReqLatestCommit          = "pullreq.approvals.require_latest_commit"
	codePullReqApprovalReqChangeRequested       = "pullreq.approvals.require_change_requested"
	codePullReqApprovalReqChangeRequestedOldSHA = "pullreq.approvals.require_change_requested_old_SHA"

	codePullReqApprovalReqCodeOwnersNoApproval       = "pullreq.approvals.require_code_owners:no_approval"
	codePullReqApprovalReqCodeOwnersChangeRequested  = "pullreq.approvals.require_code_owners:change_requested"
	codePullReqApprovalReqCodeOwnersNoLatestApproval = "pullreq.approvals.require_code_owners:no_latest_approval"

	codePullReqMergeStrategiesAllowed = "pullreq.merge.strategies_allowed"
	codePullReqMergeDeleteBranch      = "pullreq.merge.delete_branch"

	codePullReqCommentsReqResolveAll      = "pullreq.comments.require_resolve_all"
	codePullReqStatusChecksReqIdentifiers = "pullreq.status_checks.required_identifiers"
)

//nolint:gocognit // well aware of this
func (v *DefPullReq) MergeVerify(
	_ context.Context,
	in MergeVerifyInput,
) (MergeVerifyOutput, []types.RuleViolations, error) {
	var out MergeVerifyOutput
	var violations types.RuleViolations

	out.DeleteSourceBranch = v.Merge.DeleteBranch

	// pullreq.approvals

	approvedBy := make([]types.PrincipalInfo, 0, len(in.Reviewers))
	for _, reviewer := range in.Reviewers {
		switch reviewer.ReviewDecision {
		case enum.PullReqReviewDecisionApproved:
			if v.Approvals.RequireLatestCommit && reviewer.SHA != in.PullReq.SourceSHA {
				continue
			}
			approvedBy = append(approvedBy, reviewer.Reviewer)
		case enum.PullReqReviewDecisionChangeReq:
			if v.Approvals.RequireNoChangeRequest {
				if reviewer.SHA == in.PullReq.SourceSHA {
					violations.Addf(
						codePullReqApprovalReqChangeRequested,
						"Reviewer %s requested changes",
						reviewer.Reviewer.DisplayName,
					)
				} else {
					violations.Addf(
						codePullReqApprovalReqChangeRequestedOldSHA,
						"Reviewer %s requested changes for an older commit",
						reviewer.Reviewer.DisplayName,
					)
				}
			}
		case enum.PullReqReviewDecisionPending,
			enum.PullReqReviewDecisionReviewed:
		}
	}

	if len(approvedBy) < v.Approvals.RequireMinimumCount {
		if v.Approvals.RequireLatestCommit {
			violations.Addf(codePullReqApprovalReqMinCountLatest,
				"Insufficient number of approvals of the latest commit. Have %d but need at least %d.",
				len(approvedBy), v.Approvals.RequireMinimumCount)
		} else {
			violations.Addf(codePullReqApprovalReqMinCount,
				"Insufficient number of approvals. Have %d but need at least %d.",
				len(approvedBy), v.Approvals.RequireMinimumCount)
		}
	}

	if v.Approvals.RequireCodeOwners {
		for _, entry := range in.CodeOwners.EvaluationEntries {
			reviewDecision, approvers := getCodeOwnerApprovalStatus(entry)

			if reviewDecision == enum.PullReqReviewDecisionPending {
				violations.Addf(codePullReqApprovalReqCodeOwnersNoApproval,
					"Code owners approval pending for %q", entry.Pattern)
				continue
			}

			if reviewDecision == enum.PullReqReviewDecisionChangeReq {
				violations.Addf(codePullReqApprovalReqCodeOwnersChangeRequested,
					"Code owners requested changes for %q", entry.Pattern)
				continue
			}

			// pull req approved. check other settings
			if !v.Approvals.RequireLatestCommit {
				continue
			}
			latestSHAApproved := slices.ContainsFunc(approvers, func(ev codeowners.OwnerEvaluation) bool {
				return ev.ReviewSHA == in.PullReq.SourceSHA
			})
			if !latestSHAApproved {
				violations.Addf(codePullReqApprovalReqCodeOwnersNoLatestApproval,
					"Code owners approval pending on latest commit for %q", entry.Pattern)
			}
		}
	}

	// pullreq.comments

	if v.Comments.RequireResolveAll && in.PullReq.UnresolvedCount > 0 {
		violations.Addf(codePullReqCommentsReqResolveAll,
			"All comments must be resolved. There are %d unresolved comments.",
			in.PullReq.UnresolvedCount)
	}

	// pullreq.status_checks

	var violatingStatusCheckIdentifiers []string
	for _, requiredIdentifier := range v.StatusChecks.RequireIdentifiers {
		var succeeded bool
		for i := range in.CheckResults {
			if in.CheckResults[i].Identifier == requiredIdentifier {
				succeeded = in.CheckResults[i].Status == enum.CheckStatusSuccess
				break
			}
		}

		if !succeeded {
			violatingStatusCheckIdentifiers = append(violatingStatusCheckIdentifiers, requiredIdentifier)
		}
	}

	if len(violatingStatusCheckIdentifiers) > 0 {
		violations.Addf(
			codePullReqStatusChecksReqIdentifiers,
			"The following status checks are required to be completed successfully: %s",
			strings.Join(violatingStatusCheckIdentifiers, ", "),
		)
	}

	// pullreq.merge

	if in.Method == "" {
		out.AllowedMethods = enum.MergeMethods
	}

	// Note: Empty allowed strategies list means all are allowed
	if len(v.Merge.StrategiesAllowed) > 0 {
		if in.Method != "" {
			// if the Method is provided report violations if any
			if !slices.Contains(v.Merge.StrategiesAllowed, in.Method) {
				violations.Addf(codePullReqMergeStrategiesAllowed,
					"The requested merge strategy %q is not allowed. Allowed strategies are %v.",
					in.Method, v.Merge.StrategiesAllowed)
			}
		} else {
			// if the Method isn't provided return allowed strategies
			out.AllowedMethods = v.Merge.StrategiesAllowed
		}
	}

	if len(violations.Violations) > 0 {
		return out, []types.RuleViolations{violations}, nil
	}

	return out, nil, nil
}

func (v *DefPullReq) RequiredChecks(
	_ context.Context,
	_ RequiredChecksInput,
) (RequiredChecksOutput, error) {
	m := make(map[string]struct{}, len(v.StatusChecks.RequireIdentifiers))
	for _, id := range v.StatusChecks.RequireIdentifiers {
		m[id] = struct{}{}
	}
	return RequiredChecksOutput{
		RequiredIdentifiers: m,
	}, nil
}

type DefApprovals struct {
	RequireCodeOwners      bool `json:"require_code_owners,omitempty"`
	RequireMinimumCount    int  `json:"require_minimum_count,omitempty"`
	RequireLatestCommit    bool `json:"require_latest_commit,omitempty"`
	RequireNoChangeRequest bool `json:"require_no_change_request,omitempty"`
}

func (v *DefApprovals) Sanitize() error {
	if v.RequireMinimumCount < 0 {
		return errors.New("minimum count must be zero or a positive integer")
	}

	if v.RequireLatestCommit && v.RequireMinimumCount == 0 && !v.RequireCodeOwners {
		return errors.New("require latest commit can only be used with require code owners or require minimum count")
	}

	return nil
}

type DefComments struct {
	RequireResolveAll bool `json:"require_resolve_all,omitempty"`
}

func (DefComments) Sanitize() error {
	return nil
}

type DefStatusChecks struct {
	RequireIdentifiers []string `json:"require_identifiers,omitempty"`
}

// TODO [CODE-1363]: remove after identifier migration.
func (c DefStatusChecks) MarshalJSON() ([]byte, error) {
	// alias allows us to embed the original object while avoiding an infinite loop of marshaling.
	type alias DefStatusChecks
	return json.Marshal(&struct {
		alias
		RequireUIDs []string `json:"require_uids"`
	}{
		alias:       (alias)(c),
		RequireUIDs: c.RequireIdentifiers,
	})
}

// TODO [CODE-1363]: remove if we don't have any require_uids left in our DB.
func (c *DefStatusChecks) UnmarshalJSON(data []byte) error {
	// alias allows us to extract the original object while avoiding an infinite loop of unmarshaling.
	type alias DefStatusChecks
	res := struct {
		alias
		RequireUIDs []string `json:"require_uids"`
	}{}

	err := json.Unmarshal(data, &res)
	if err != nil {
		return fmt.Errorf("failed to unmarshal to alias type with required uids: %w", err)
	}

	*c = DefStatusChecks(res.alias)
	if len(c.RequireIdentifiers) == 0 {
		c.RequireIdentifiers = res.RequireUIDs
	}

	return nil
}

func (c *DefStatusChecks) Sanitize() error {
	if err := validateIdentifierSlice(c.RequireIdentifiers); err != nil {
		return fmt.Errorf("required identifiers error: %w", err)
	}

	return nil
}

type DefMerge struct {
	StrategiesAllowed []enum.MergeMethod `json:"strategies_allowed,omitempty"`
	DeleteBranch      bool               `json:"delete_branch,omitempty"`
}

func (v *DefMerge) Sanitize() error {
	m := make(map[enum.MergeMethod]struct{}, 0)
	for _, strategy := range v.StrategiesAllowed {
		if _, ok := strategy.Sanitize(); !ok {
			return fmt.Errorf("unrecognized merge strategy: %s", strategy)
		}

		if _, ok := m[strategy]; ok {
			return fmt.Errorf("duplicate entry in merge strategy list: %s", strategy)
		}

		m[strategy] = struct{}{}
	}

	slices.Sort(v.StrategiesAllowed)

	return nil
}

type DefPush struct {
	Block bool `json:"block,omitempty"`
}

func (v *DefPush) Sanitize() error {
	return nil
}

type DefPullReq struct {
	Approvals    DefApprovals    `json:"approvals"`
	Comments     DefComments     `json:"comments"`
	StatusChecks DefStatusChecks `json:"status_checks"`
	Merge        DefMerge        `json:"merge"`
}

func (v *DefPullReq) Sanitize() error {
	if err := v.Approvals.Sanitize(); err != nil {
		return fmt.Errorf("approvals: %w", err)
	}

	if err := v.Comments.Sanitize(); err != nil {
		return fmt.Errorf("comments: %w", err)
	}

	if err := v.StatusChecks.Sanitize(); err != nil {
		return fmt.Errorf("status checks: %w", err)
	}

	if err := v.Merge.Sanitize(); err != nil {
		return fmt.Errorf("merge: %w", err)
	}

	return nil
}

func getCodeOwnerApprovalStatus(
	entry codeowners.EvaluationEntry,
) (enum.PullReqReviewDecision, []codeowners.OwnerEvaluation) {
	approvers := make([]codeowners.OwnerEvaluation, 0)

	// users
	for _, o := range entry.OwnerEvaluations {
		if o.ReviewDecision == enum.PullReqReviewDecisionChangeReq {
			return enum.PullReqReviewDecisionChangeReq, nil
		}
		if o.ReviewDecision == enum.PullReqReviewDecisionApproved {
			approvers = append(approvers, o)
		}
	}

	// usergroups
	for _, u := range entry.UserGroupOwnerEvaluations {
		for _, o := range u.Evaluations {
			if o.ReviewDecision == enum.PullReqReviewDecisionChangeReq {
				return enum.PullReqReviewDecisionChangeReq, nil
			}
			if o.ReviewDecision == enum.PullReqReviewDecisionApproved {
				approvers = append(approvers, o)
			}
		}
	}

	if len(approvers) > 0 {
		return enum.PullReqReviewDecisionApproved, approvers
	}
	return enum.PullReqReviewDecisionPending, nil
}
