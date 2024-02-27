/*
 * Copyright 2023 Harness, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { ButtonProps, Container, FlexExpander, Layout, Select, SelectOption, Text, useToaster } from '@harnessio/uicore'
import { useGet, useMutate } from 'restful-react'
import { orderBy } from 'lodash-es'
import type { GitInfoProps } from 'utils/GitUtils'
import { useStrings } from 'framework/strings'
import { useAppContext } from 'AppContext'
import type { TypesPullReqActivity, TypesPullReq, TypesPullReqStats, TypesCodeOwnerEvaluation } from 'services/code'
import { CommentAction, CommentBox, CommentBoxOutletPosition, CommentItem } from 'components/CommentBox/CommentBox'
import { useConfirmAct } from 'hooks/useConfirmAction'
import { getErrorMessage, orderSortDate, ButtonRoleProps } from 'utils/Utils'
import { activityToCommentItem } from 'components/DiffViewer/DiffViewerUtils'
import { NavigationCheck } from 'components/NavigationCheck/NavigationCheck'
import { ThreadSection } from 'components/ThreadSection/ThreadSection'
import { CodeCommentStatusSelect } from 'components/CodeCommentStatusSelect/CodeCommentStatusSelect'
import { CodeCommentStatusButton } from 'components/CodeCommentStatusButton/CodeCommentStatusButton'
import { CodeCommentSecondarySaveButton } from 'components/CodeCommentSecondarySaveButton/CodeCommentSecondarySaveButton'
import type { PRChecksDecisionResult } from 'hooks/usePRChecksDecision'
import { UserPreference, useUserPreference } from 'hooks/useUserPreference'
import { PullRequestTabContentWrapper } from '../PullRequestTabContentWrapper'
import { DescriptionBox } from './DescriptionBox'
import { PullRequestActionsBox } from './PullRequestActionsBox/PullRequestActionsBox'
import PullRequestSideBar from './PullRequestSideBar/PullRequestSideBar'
import { isCodeComment, isComment, isSystemComment } from '../PullRequestUtils'
import { ChecksOverview } from '../Checks/ChecksOverview'
import { usePullReqActivities } from '../useGetPullRequestInfo'
import { CodeCommentHeader } from './CodeCommentHeader'
import { SystemComment } from './SystemComment'
import CodeOwnersOverview from '../CodeOwners/CodeOwnersOverview'
import css from './Conversation.module.scss'

export interface ConversationProps extends Pick<GitInfoProps, 'repoMetadata' | 'pullReqMetadata'> {
  prStats?: TypesPullReqStats
  showEditDescription?: boolean
  onCancelEditDescription: () => void
  onDescriptionSaved: () => void
  prChecksDecisionResult?: PRChecksDecisionResult
  standalone: boolean
  routingId: string
}

export const Conversation: React.FC<ConversationProps> = ({
  repoMetadata,
  pullReqMetadata,
  onDescriptionSaved,
  prStats,
  showEditDescription,
  onCancelEditDescription,
  prChecksDecisionResult,
  standalone,
  routingId
}) => {
  const { getString } = useStrings()
  const { currentUser } = useAppContext()
  const activities = usePullReqActivities()
  const { data: reviewers, refetch: refetchReviewers } = useGet<Unknown[]>({
    path: `/api/v1/repos/${repoMetadata.path}/+/pullreq/${pullReqMetadata.number}/reviewers`,
    debounce: 500
  })

  const { data: codeOwners, refetch: refetchCodeOwners } = useGet<TypesCodeOwnerEvaluation>({
    path: `/api/v1/repos/${repoMetadata.path}/+/pullreq/${pullReqMetadata.number}/codeowners`,
    debounce: 500
  })

  const { showError } = useToaster()
  const [dateOrderSort, setDateOrderSort] = useUserPreference<orderSortDate.ASC | orderSortDate.DESC>(
    UserPreference.PULL_REQUEST_ACTIVITY_ORDER,
    orderSortDate.ASC
  )
  const activityFilters = useActivityFilters()
  const [activityFilter, setActivityFilter] = useUserPreference<SelectOption>(
    UserPreference.PULL_REQUEST_ACTIVITY_FILTER,
    activityFilters[0] as SelectOption
  )

  const activityBlocks = useMemo(() => {
    // Each block may have one or more activities which are grouped into it. For example, one comment block
    // contains a parent comment and multiple replied comments
    const blocks: CommentItem<TypesPullReqActivity>[][] = []

    // Determine all parent activities
    const parentActivities = orderBy(
      activities?.filter(activity => !activity.parent_id) || [],
      'created',
      dateOrderSort
    ).map(_comment => [_comment])

    // Then add their children as follow-up elements (same array)
    parentActivities?.forEach(parentActivity => {
      const childActivities = activities?.filter(activity => activity.parent_id === parentActivity[0].id)

      childActivities?.forEach(childComment => {
        parentActivity.push(childComment)
      })
    })

    parentActivities?.forEach(parentActivity => {
      blocks.push(parentActivity.map(activityToCommentItem))
    })

    switch (activityFilter.value) {
      case PRCommentFilterType.ALL_COMMENTS:
        return blocks.filter(_activities => !isSystemComment(_activities))

      case PRCommentFilterType.RESOLVED_COMMENTS:
        return blocks.filter(
          _activities => _activities[0].payload?.resolved && (isCodeComment(_activities) || isComment(_activities))
        )

      case PRCommentFilterType.UNRESOLVED_COMMENTS:
        return blocks.filter(
          _activities => !_activities[0].payload?.resolved && (isComment(_activities) || isCodeComment(_activities))
        )

      case PRCommentFilterType.MY_COMMENTS: {
        const allCommentBlock = blocks.filter(_activities => !isSystemComment(_activities))
        const userCommentsOnly = allCommentBlock.filter(_activities => {
          const userCommentReply = _activities.filter(
            authorIsUser => currentUser?.uid && authorIsUser.payload?.author?.uid === currentUser?.uid
          )
          return userCommentReply.length !== 0
        })
        return userCommentsOnly
      }
    }

    return blocks
  }, [activities, dateOrderSort, activityFilter, currentUser?.uid])
  const path = useMemo(
    () => `/api/v1/repos/${repoMetadata.path}/+/pullreq/${pullReqMetadata.number}/comments`,
    [repoMetadata.path, pullReqMetadata.number]
  )
  const { mutate: saveComment } = useMutate({ verb: 'POST', path })
  const { mutate: updateComment } = useMutate({ verb: 'PATCH', path: ({ id }) => `${path}/${id}` })
  const { mutate: deleteComment } = useMutate({ verb: 'DELETE', path: ({ id }) => `${path}/${id}` })
  const confirmAct = useConfirmAct()
  const [dirtyNewComment, setDirtyNewComment] = useState(false)
  const [dirtyCurrentComments, setDirtyCurrentComments] = useState(false)
  const onPRStateChanged = useCallback(() => {
    refetchCodeOwners()
  }, [refetchCodeOwners])
  const hasDescription = useMemo(() => !!pullReqMetadata?.description?.length, [pullReqMetadata?.description?.length])

  useEffect(() => {
    if (prStats) {
      refetchCodeOwners()
    }
  }, [
    prStats,
    prStats?.conversations,
    prStats?.unresolved_count,
    pullReqMetadata?.title,
    pullReqMetadata?.state,
    pullReqMetadata?.source_sha,
    refetchCodeOwners
  ])

  const newCommentBox = useMemo(() => {
    return (
      <CommentBox
        routingId={routingId}
        standalone={standalone}
        repoMetadata={repoMetadata}
        fluid
        boxClassName={css.commentBox}
        editorClassName={css.commentEditor}
        commentItems={[]}
        currentUserName={currentUser?.display_name || currentUser?.email || ''}
        resetOnSave
        hideCancel={false}
        setDirty={setDirtyNewComment}
        enableReplyPlaceHolder={true}
        autoFocusAndPosition={true}
        handleAction={async (_action, value) => {
          let result = true
          let updatedItem: CommentItem<TypesPullReqActivity> | undefined = undefined

          await saveComment({ text: value })
            .then((newComment: TypesPullReqActivity) => {
              updatedItem = activityToCommentItem(newComment)
            })
            .catch(exception => {
              result = false
              showError(getErrorMessage(exception), 0)
            })

          return [result, updatedItem]
        }}
      />
    ) // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentUser, saveComment, showError])

  const renderedActivityBlocks = useMemo(
    () =>
      activityBlocks?.map((commentItems, index) => {
        const threadId = commentItems[0].payload?.id

        if (isSystemComment(commentItems)) {
          return (
            <ThreadSection
              key={`thread-${threadId}`}
              onlyTitle
              lastItem={activityBlocks.length - 1 === index}
              title={
                <SystemComment
                  key={`system-${threadId}`}
                  pullReqMetadata={pullReqMetadata}
                  commentItems={commentItems}
                  repoMetadataPath={repoMetadata.path}
                />
              }></ThreadSection>
          )
        }
        return (
          <ThreadSection
            key={`comment-${threadId}`}
            onlyTitle={
              activityBlocks[index + 1] !== undefined && isSystemComment(activityBlocks[index + 1]) ? true : false
            }
            inCommentBox={
              activityBlocks[index + 1] !== undefined && isSystemComment(activityBlocks[index + 1]) ? true : false
            }
            title={
              <CommentBox
                routingId={routingId}
                standalone={standalone}
                repoMetadata={repoMetadata}
                key={threadId}
                fluid
                boxClassName={css.threadbox}
                outerClassName={css.hideDottedLine}
                commentItems={commentItems}
                currentUserName={currentUser?.display_name || currentUser?.email || ''}
                setDirty={setDirtyCurrentComments}
                enableReplyPlaceHolder={true}
                autoFocusAndPosition={true}
                handleAction={async (action, value, commentItem) => {
                  let result = true
                  let updatedItem: CommentItem<TypesPullReqActivity> | undefined = undefined
                  const id = (commentItem as CommentItem<TypesPullReqActivity>)?.payload?.id

                  switch (action) {
                    case CommentAction.DELETE:
                      result = false
                      await confirmAct({
                        message: getString('deleteCommentConfirm'),
                        action: async () => {
                          await deleteComment({}, { pathParams: { id } })
                            .then(() => {
                              result = true
                            })
                            .catch(exception => {
                              result = false
                              showError(getErrorMessage(exception), 0, getString('pr.failedToDeleteComment'))
                            })
                        }
                      })
                      break

                    case CommentAction.REPLY:
                      await saveComment({ text: value, parent_id: Number(threadId) })
                        .then(newComment => {
                          updatedItem = activityToCommentItem(newComment)
                        })
                        .catch(exception => {
                          result = false
                          showError(getErrorMessage(exception), 0, getString('pr.failedToSaveComment'))
                        })
                      break

                    case CommentAction.UPDATE:
                      await updateComment({ text: value }, { pathParams: { id } })
                        .then(newComment => {
                          updatedItem = activityToCommentItem(newComment)
                        })
                        .catch(exception => {
                          result = false
                          showError(getErrorMessage(exception), 0, getString('pr.failedToSaveComment'))
                        })
                      break
                  }

                  return [result, updatedItem]
                }}
                outlets={{
                  [CommentBoxOutletPosition.TOP_OF_FIRST_COMMENT]: isCodeComment(commentItems) && (
                    <CodeCommentHeader
                      commentItems={commentItems}
                      threadId={threadId}
                      repoMetadata={repoMetadata}
                      pullReqMetadata={pullReqMetadata}
                    />
                  ),
                  [CommentBoxOutletPosition.LEFT_OF_OPTIONS_MENU]: (
                    <CodeCommentStatusSelect
                      repoMetadata={repoMetadata}
                      pullReqMetadata={pullReqMetadata}
                      commentItems={commentItems}
                    />
                  ),
                  [CommentBoxOutletPosition.RIGHT_OF_REPLY_PLACEHOLDER]: (
                    <CodeCommentStatusButton
                      repoMetadata={repoMetadata}
                      pullReqMetadata={pullReqMetadata}
                      commentItems={commentItems}
                    />
                  ),
                  [CommentBoxOutletPosition.BETWEEN_SAVE_AND_CANCEL_BUTTONS]: (props: ButtonProps) => (
                    <CodeCommentSecondarySaveButton
                      repoMetadata={repoMetadata}
                      pullReqMetadata={pullReqMetadata as TypesPullReq}
                      commentItems={commentItems}
                      {...props}
                    />
                  )
                }}
              />
            }></ThreadSection>
        )
      }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [activityBlocks, currentUser, pullReqMetadata]
  )

  return (
    <PullRequestTabContentWrapper>
      <Container>
        <Layout.Vertical spacing="xlarge">
          <PullRequestActionsBox
            repoMetadata={repoMetadata}
            pullReqMetadata={pullReqMetadata}
            onPRStateChanged={onPRStateChanged}
            refetchReviewers={refetchReviewers}
          />
          <Container>
            <Layout.Horizontal width="calc(var(--page-container-width) - 48px)">
              <Container width={`70%`}>
                <Layout.Vertical spacing="xlarge">
                  {prChecksDecisionResult && (
                    <ChecksOverview
                      repoMetadata={repoMetadata}
                      pullReqMetadata={pullReqMetadata}
                      prChecksDecisionResult={prChecksDecisionResult}
                      codeOwners={codeOwners as TypesCodeOwnerEvaluation}
                    />
                  )}
                  {codeOwners && prChecksDecisionResult && (
                    <CodeOwnersOverview
                      standalone={standalone}
                      codeOwners={codeOwners}
                      repoMetadata={repoMetadata}
                      pullReqMetadata={pullReqMetadata}
                      prChecksDecisionResult={prChecksDecisionResult}
                    />
                  )}

                  {(hasDescription || showEditDescription) && (
                    <DescriptionBox
                      routingId={routingId}
                      standalone={standalone}
                      repoMetadata={repoMetadata}
                      pullReqMetadata={pullReqMetadata}
                      onDescriptionSaved={onDescriptionSaved}
                      onCancelEditDescription={onCancelEditDescription}
                      prStats={prStats}
                    />
                  )}

                  <Layout.Horizontal
                    className={css.sortContainer}
                    padding={{ top: hasDescription || showEditDescription ? 'xxlarge' : undefined, bottom: 'medium' }}>
                    <Container>
                      <Select
                        items={activityFilters}
                        value={activityFilter}
                        className={css.selectButton}
                        onChange={newState => {
                          setActivityFilter(newState)
                        }}
                      />
                    </Container>
                    <FlexExpander />
                    <Text
                      {...ButtonRoleProps}
                      className={css.timeButton}
                      rightIconProps={{ size: 24 }}
                      rightIcon={dateOrderSort === orderSortDate.ASC ? 'code-ascending' : 'code-descending'}
                      onClick={() => {
                        if (dateOrderSort === orderSortDate.ASC) {
                          setDateOrderSort(orderSortDate.DESC)
                        } else {
                          setDateOrderSort(orderSortDate.ASC)
                        }
                      }}>
                      {dateOrderSort === orderSortDate.ASC ? getString('ascending') : getString('descending')}
                    </Text>
                  </Layout.Horizontal>

                  {dateOrderSort != orderSortDate.DESC ? null : (
                    <Container className={css.descContainer}>{newCommentBox}</Container>
                  )}

                  {renderedActivityBlocks}

                  {dateOrderSort != orderSortDate.ASC ? null : (
                    <Container className={css.ascContainer}>{newCommentBox}</Container>
                  )}
                </Layout.Vertical>
              </Container>

              <PullRequestSideBar
                reviewers={reviewers}
                repoMetadata={repoMetadata}
                pullRequestMetadata={pullReqMetadata}
                refetchReviewers={refetchReviewers}
              />
            </Layout.Horizontal>
          </Container>
        </Layout.Vertical>
      </Container>
      <NavigationCheck when={dirtyCurrentComments || dirtyNewComment} />
    </PullRequestTabContentWrapper>
  )
}

export enum PRCommentFilterType {
  SHOW_EVERYTHING = 'showEverything',
  ALL_COMMENTS = 'allComments',
  MY_COMMENTS = 'myComments',
  RESOLVED_COMMENTS = 'resolvedComments',
  UNRESOLVED_COMMENTS = 'unresolvedComments'
}

function useActivityFilters() {
  const { getString } = useStrings()

  return useMemo(
    () => [
      {
        label: getString('showEverything'),
        value: PRCommentFilterType.SHOW_EVERYTHING
      },
      {
        label: getString('allComments'),
        value: PRCommentFilterType.ALL_COMMENTS
      },
      {
        label: getString('myComments'),
        value: PRCommentFilterType.MY_COMMENTS
      },
      {
        label: getString('unrsolvedComment'),
        value: PRCommentFilterType.UNRESOLVED_COMMENTS
      },
      {
        label: getString('resolvedComments'),
        value: PRCommentFilterType.RESOLVED_COMMENTS
      }
    ],
    [getString]
  )
}
