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

import React, { useMemo } from 'react'
import { Falsy, Match, Render, Truthy } from 'react-jsx-match'
import {
  Container,
  Text,
  useToggle,
  Button,
  ButtonVariation,
  ButtonSize,
  Utils,
  TableV2,
  Layout,
  Avatar,
  stringSubstitute
} from '@harnessio/uicore'
import cx from 'classnames'
import { Color, FontVariation } from '@harnessio/design-system'
import type { CellProps, Column } from 'react-table'
import type { GitInfoProps } from 'utils/GitUtils'
import { useStrings } from 'framework/strings'
import { ExecutionState, ExecutionStatus } from 'components/ExecutionStatus/ExecutionStatus'
import { useShowRequestError } from 'hooks/useShowRequestError'
import type { TypesCodeOwnerEvaluation, TypesCodeOwnerEvaluationEntry } from 'services/code'
import type { PRChecksDecisionResult } from 'hooks/usePRChecksDecision'
import { findChangeReqDecisions, findWaitingDecisions } from 'utils/Utils'
import css from './CodeOwnersOverview.module.scss'

interface ChecksOverviewProps extends Pick<GitInfoProps, 'repoMetadata' | 'pullReqMetadata'> {
  prChecksDecisionResult: PRChecksDecisionResult
  codeOwners?: TypesCodeOwnerEvaluation
  standalone: boolean
}

enum CodeOwnerReqDecision {
  CHANGEREQ = 'changereq',
  APPROVED = 'approved',
  WAIT_FOR_APPROVAL = ''
}

export function CodeOwnersOverview({
  codeOwners,
  repoMetadata,
  pullReqMetadata,
  prChecksDecisionResult,
  standalone
}: ChecksOverviewProps) {
  const { getString } = useStrings()
  const [isExpanded, toggleExpanded] = useToggle(false)
  const { error } = prChecksDecisionResult

  useShowRequestError(error)

  const changeReqEntries = findChangeReqDecisions(codeOwners?.evaluation_entries, CodeOwnerReqDecision.CHANGEREQ)
  const waitingEntries = findWaitingDecisions(codeOwners?.evaluation_entries)

  const approvalEntries = findChangeReqDecisions(codeOwners?.evaluation_entries, CodeOwnerReqDecision.APPROVED)

  const checkEntries = (
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    changeReqArr: any[], // eslint-disable-next-line @typescript-eslint/no-explicit-any
    waitingEntriesArr: any[], // eslint-disable-next-line @typescript-eslint/no-explicit-any
    approvalEntriesArr: any[]
  ): { borderColor: string; message: string; overallStatus: ExecutionState } => {
    if (changeReqArr.length !== 0) {
      return {
        borderColor: 'red800',
        overallStatus: ExecutionState.FAILURE,
        message: stringSubstitute(getString('codeOwner.changesRequested'), { count: changeReqArr.length }) as string
      }
    } else if (waitingEntriesArr.length !== 0) {
      return {
        borderColor: 'orange800',
        message: stringSubstitute(getString('codeOwner.waitToApprove'), { count: waitingEntriesArr.length }) as string,
        overallStatus: ExecutionState.PENDING
      }
    }
    return {
      borderColor: 'green800',
      message: stringSubstitute(getString('codeOwner.approvalCompleted'), {
        count: approvalEntriesArr.length || '0',
        total: codeOwners?.evaluation_entries?.length
      }) as string,
      overallStatus: ExecutionState.SUCCESS
    }
  }
  const { borderColor, message, overallStatus } = checkEntries(changeReqEntries, waitingEntries, approvalEntries)
  return codeOwners?.evaluation_entries?.length ? (
    <Container
      className={cx(css.main, { [css.codeOwner]: !standalone })}
      margin={{ top: 'medium', bottom: pullReqMetadata.description ? undefined : 'large' }}
      style={{ '--border-color': Utils.getRealCSSColor(borderColor) } as React.CSSProperties}>
      <Match expr={isExpanded}>
        <Truthy>
          {codeOwners && (
            <CodeOwnerSections repoMetadata={repoMetadata} pullReqMetadata={pullReqMetadata} data={codeOwners} />
          )}
        </Truthy>
        <Falsy>
          <Layout.Horizontal spacing="small" className={css.layout}>
            <ExecutionStatus inExecution={true} isCi={true} status={overallStatus} noBackground iconOnly />
            <Text font={{ variation: FontVariation.LEAD }}>{getString('codeOwner.title')}</Text>
            <Text
              color={borderColor}
              padding={{ left: 'small' }}
              font={{ variation: FontVariation.FORM_MESSAGE_WARNING }}>
              {message}
            </Text>
          </Layout.Horizontal>
        </Falsy>
      </Match>
      <Button
        className={css.showMore}
        variation={ButtonVariation.LINK}
        size={ButtonSize.SMALL}
        text={getString(isExpanded ? 'showLess' : 'showMore')}
        rightIcon={isExpanded ? 'main-chevron-up' : 'main-chevron-down'}
        iconProps={{ size: 10, margin: { left: 'xsmall' } }}
        onClick={toggleExpanded}
      />
    </Container>
  ) : null
}

interface CodeOwnerSectionsProps extends Pick<GitInfoProps, 'repoMetadata' | 'pullReqMetadata'> {
  data: TypesCodeOwnerEvaluation
}

const CodeOwnerSections: React.FC<CodeOwnerSectionsProps> = ({ repoMetadata, pullReqMetadata, data }) => {
  return (
    <Container className={css.checks}>
      <Layout.Vertical spacing="medium">
        <CodeOwnerSection repoMetadata={repoMetadata} pullReqMetadata={pullReqMetadata} data={data} />
      </Layout.Vertical>
    </Container>
  )
}

const CodeOwnerSection: React.FC<CodeOwnerSectionsProps> = ({ data }) => {
  const { getString } = useStrings()
  const columns = useMemo(
    () =>
      [
        {
          id: 'CODE',
          width: '50%',
          sort: true,
          Header: 'CODE',
          accessor: 'CODE',
          Cell: ({ row }: CellProps<TypesCodeOwnerEvaluationEntry>) => {
            return (
              <Text lineClamp={1} padding={{ left: 'small', right: 'small' }} color={Color.BLACK}>
                {row.original.pattern}
              </Text>
            )
          }
        },
        {
          id: 'Owners',
          width: '20%',
          sort: true,
          Header: 'OWNERS',
          accessor: 'OWNERS',
          Cell: ({ row }: CellProps<TypesCodeOwnerEvaluationEntry>) => {
            return (
              <Layout.Horizontal
                key={`keyContainer-${row.original.pattern}`}
                className={css.ownerContainer}
                spacing="tiny">
                {row.original.owner_evaluations?.map(({ owner }, idx) => {
                  if (idx < 4) {
                    return (
                      <Avatar
                        key={`text-${owner?.display_name}-${idx}-avatar`}
                        hoverCard={true}
                        email={owner?.email || ' '}
                        size="small"
                        name={owner?.display_name || ''}
                      />
                    )
                  }
                  if (
                    idx === 4 &&
                    row.original.owner_evaluations?.length &&
                    row.original.owner_evaluations?.length > 4
                  ) {
                    return (
                      <Text
                        key={`text-${owner?.display_name}-${idx}-top`}
                        padding={{ top: 'xsmall' }}
                        tooltipProps={{ isDark: true }}
                        tooltip={
                          <Container width={215} padding={'small'}>
                            <Layout.Horizontal key={`tooltip-${idx}`} className={css.ownerTooltip}>
                              {row.original.owner_evaluations?.map((entry, entryidx) => (
                                <Text
                                  key={`text-${entry.owner?.display_name}-${entryidx}`}
                                  lineClamp={1}
                                  color={Color.GREY_0}
                                  padding={{ right: 'small' }}>
                                  {row.original.owner_evaluations?.length === entryidx + 1
                                    ? `${entry.owner?.display_name}`
                                    : `${entry.owner?.display_name}, `}
                                </Text>
                              ))}
                            </Layout.Horizontal>
                          </Container>
                        }
                        flex={{ alignItems: 'center' }}>{`+${row.original.owner_evaluations?.length - 4}`}</Text>
                    )
                  }
                  return null
                })}
              </Layout.Horizontal>
            )
          }
        },
        {
          id: 'approvals',
          Header: 'APPROVALS',
          width: '15%',
          sort: true,
          accessor: 'APPROVALS',
          Cell: ({ row }: CellProps<TypesCodeOwnerEvaluationEntry>) => {
            const approvedEvaluations = row?.original?.owner_evaluations?.filter(
              evaluation => evaluation.review_decision === 'approved'
            )
            const changeReqEvaluations = row?.original?.owner_evaluations?.filter(
              evaluation => evaluation.review_decision === 'changereq'
            )
            if (changeReqEvaluations && changeReqEvaluations.length !== 0) {
              return (
                <Text
                  className={css.approvalText}
                  icon="warning-sign"
                  iconProps={{ color: Color.RED_700, size: 13 }}
                  color={Color.RED_700}>
                  {getString('requestChanges')}
                </Text>
              )
            }
            if (approvedEvaluations && approvedEvaluations.length !== 0) {
              return (
                <Text
                  className={css.approvalText}
                  icon={'execution-success'}
                  iconProps={{ color: Color.GREEN_700, size: 13 }}
                  color={Color.GREEN_700}>
                  {getString('approved')}
                </Text>
              )
            }
            return (
              <Text flex className={cx(css.approvalText, css.waitingContainer)} color={Color.ORANGE_700}>
                <Container className={css.circle}></Container>
                {getString('pending')}
              </Text>
            )
          }
        },
        {
          id: 'approvedBy',
          Header: 'APPROVED BY',
          sort: true,
          width: '15%',
          accessor: 'APPROVED BY',
          Cell: ({ row }: CellProps<TypesCodeOwnerEvaluationEntry>) => {
            const approvedEvaluations = row?.original?.owner_evaluations?.filter(
              evaluation => evaluation.review_decision === 'approved'
            )
            return (
              <Layout.Horizontal className={css.ownerContainer} spacing="tiny">
                {approvedEvaluations?.map(({ owner }, idx) => {
                  if (idx < 4) {
                    return (
                      <Avatar
                        key={`approved-${owner?.display_name}-avatar`}
                        hoverCard={true}
                        email={owner?.email || ' '}
                        size="small"
                        name={owner?.display_name || ''}
                      />
                    )
                  }
                  if (idx === 4 && approvedEvaluations.length && approvedEvaluations.length > 4) {
                    return (
                      <Text
                        key={`approved-${owner?.display_name}-text`}
                        padding={{ top: 'xsmall' }}
                        tooltipProps={{ isDark: true }}
                        tooltip={
                          <Container width={215} padding={'small'}>
                            <Layout.Horizontal className={css.ownerTooltip}>
                              {approvedEvaluations?.map(entry => (
                                <Text
                                  key={`approved-${entry.owner?.display_name}`}
                                  lineClamp={1}
                                  color={Color.GREY_0}
                                  padding={{ right: 'small' }}>{`${entry.owner?.display_name}, `}</Text>
                              ))}
                            </Layout.Horizontal>
                          </Container>
                        }
                        flex={{ alignItems: 'center' }}>{`+${approvedEvaluations.length - 4}`}</Text>
                    )
                  }
                  return null
                })}
              </Layout.Horizontal>
            )
          }
        }
      ] as unknown as Column<TypesCodeOwnerEvaluationEntry>[], // eslint-disable-next-line react-hooks/exhaustive-deps
    []
  )
  return (
    <Render when={data?.evaluation_entries?.length}>
      <Container>
        <Layout.Vertical spacing="small">
          <Text padding={{ left: 'medium' }} font={{ variation: FontVariation.SMALL_BOLD }}>
            {getString('codeOwner.title')}
          </Text>

          <TableV2
            className={css.codeOwnerTable}
            sortable
            columns={columns}
            data={data?.evaluation_entries as TypesCodeOwnerEvaluationEntry[]}
            getRowClassName={() => css.row}
          />
        </Layout.Vertical>
      </Container>
    </Render>
  )
}

export default CodeOwnersOverview
