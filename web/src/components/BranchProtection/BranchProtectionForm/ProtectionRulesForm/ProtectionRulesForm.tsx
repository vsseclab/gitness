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
import React from 'react'

import cx from 'classnames'
import { Container, FlexExpander, FormInput, Layout, SelectOption, Text } from '@harnessio/uicore'
import { Icon } from '@harnessio/icons'

import { FontVariation } from '@harnessio/design-system'
import { useStrings } from 'framework/strings'
import css from '../BranchProtectionForm.module.scss'
const ProtectionRulesForm = (props: {
  requireStatusChecks: boolean
  minReviewers: boolean
  statusOptions: SelectOption[]
  statusChecks: string[]
  limitMergeStrats: boolean // eslint-disable-next-line @typescript-eslint/no-explicit-any
  setFieldValue: (field: string, value: any, shouldValidate?: boolean | undefined) => void
  setSearchStatusTerm: React.Dispatch<React.SetStateAction<string>>
}) => {
  const {
    setFieldValue,
    statusChecks,
    setSearchStatusTerm,
    minReviewers,
    requireStatusChecks,
    statusOptions,
    limitMergeStrats
  } = props
  const { getString } = useStrings()
  const filteredStatusOptions = statusOptions.filter(
    (item: SelectOption) => !statusChecks.includes(item.value as string)
  )
  return (
    <Container margin={{ top: 'medium' }} className={css.generalContainer}>
      <Text className={css.headingSize} padding={{ bottom: 'medium' }} font={{ variation: FontVariation.H4 }}>
        {getString('branchProtection.protectionSelectAll')}
      </Text>

      <FormInput.CheckBox
        className={css.checkboxLabel}
        label={getString('branchProtection.blockBranchCreation')}
        name={'blockBranchCreation'}
      />
      <Text padding={{ left: 'xlarge' }} className={css.checkboxText}>
        {getString('branchProtection.blockBranchCreationText')}
      </Text>

      <hr className={css.dividerContainer} />
      <FormInput.CheckBox
        className={css.checkboxLabel}
        label={getString('branchProtection.blockBranchDeletion')}
        name={'blockBranchDeletion'}
      />
      <Text padding={{ left: 'xlarge' }} className={css.checkboxText}>
        {getString('branchProtection.blockBranchDeletionText')}
      </Text>
      <hr className={css.dividerContainer} />
      <FormInput.CheckBox
        className={css.checkboxLabel}
        label={getString('branchProtection.requirePr')}
        name={'requirePr'}
      />
      <Text padding={{ left: 'xlarge' }} className={css.checkboxText}>
        {getString('branchProtection.requirePrText')}
      </Text>

      <hr className={css.dividerContainer} />
      <FormInput.CheckBox
        className={css.checkboxLabel}
        label={getString('branchProtection.requireMinReviewersTitle')}
        name={'requireMinReviewers'}
        onChange={e => {
          if ((e.target as HTMLInputElement).checked) {
            setFieldValue('minReviewers', 1)
          }
        }}
      />
      <Text padding={{ left: 'xlarge' }} className={css.checkboxText}>
        {getString('branchProtection.requireMinReviewersContent')}
      </Text>
      {minReviewers && (
        <Container padding={{ left: 'xlarge', top: 'medium' }}>
          <FormInput.Text
            className={cx(css.widthContainer, css.minText)}
            name={'minReviewers'}
            placeholder={getString('branchProtection.minNumberPlaceholder')}
            label={getString('branchProtection.minNumber')}
          />
        </Container>
      )}
      <hr className={css.dividerContainer} />
      <FormInput.CheckBox
        className={css.checkboxLabel}
        label={getString('branchProtection.reqReviewFromCodeOwnerTitle')}
        name={'requireCodeOwner'}
      />
      <Text padding={{ left: 'xlarge' }} className={css.checkboxText}>
        {getString('branchProtection.reqReviewFromCodeOwnerText')}
      </Text>

      <hr className={css.dividerContainer} />
      <FormInput.CheckBox
        className={css.checkboxLabel}
        label={getString('branchProtection.reqNewChangesTitle')}
        name={'requireNewChanges'}
      />
      <Text padding={{ left: 'xlarge' }} className={css.checkboxText}>
        {getString('branchProtection.reqNewChangesText')}
      </Text>

      <hr className={css.dividerContainer} />
      <FormInput.CheckBox
        className={css.checkboxLabel}
        label={getString('branchProtection.reqCommentResolutionTitle')}
        name={'requireCommentResolution'}
      />
      <Text padding={{ left: 'xlarge' }} className={css.checkboxText}>
        {getString('branchProtection.reqCommentResolutionText')}
      </Text>

      <hr className={css.dividerContainer} />
      <FormInput.CheckBox
        className={css.checkboxLabel}
        label={getString('branchProtection.reqStatusChecksTitle')}
        name={'requireStatusChecks'}
      />
      <Text padding={{ left: 'xlarge' }} className={css.checkboxText}>
        {getString('branchProtection.reqStatusChecksText')}
      </Text>
      {requireStatusChecks && (
        <Container padding={{ left: 'xlarge', top: 'large' }} className={css.widthContainer}>
          <FormInput.Select
            onQueryChange={setSearchStatusTerm}
            items={filteredStatusOptions}
            value={{ label: '', value: '' }}
            placeholder={getString('selectStatuses')}
            onChange={item => {
              statusChecks?.push(item.value as string)
              const uniqueArr = Array.from(new Set(statusChecks))
              setFieldValue('statusChecks', uniqueArr)
            }}
            label={getString('branchProtection.statusCheck')}
            name={'statusSelect'}></FormInput.Select>
        </Container>
      )}
      {requireStatusChecks && (
        <Container padding={{ left: 'xlarge' }}>
          <Container className={cx(css.statusWidthContainer, css.bypassContainer)}>
            {statusChecks?.map((status, idx) => {
              return (
                <Layout.Horizontal key={`${idx}-${status[1]}`} flex={{ align: 'center-center' }} padding={'small'}>
                  {/* <Avatar hoverCard={false} size="small" name={status.value as string} /> */}
                  <Text padding={{ top: 'tiny' }} lineClamp={1}>
                    {status}
                  </Text>
                  <FlexExpander />
                  <Icon
                    className={css.codeClose}
                    name="code-close"
                    onClick={() => {
                      const filteredData = statusChecks.filter(item => !(item === status))
                      setFieldValue('statusChecks', filteredData)
                    }}
                  />
                </Layout.Horizontal>
              )
            })}
          </Container>
        </Container>
      )}
      <hr className={css.dividerContainer} />
      <FormInput.CheckBox
        className={css.checkboxLabel}
        label={getString('branchProtection.limitMergeStrategies')}
        name={'limitMergeStrategies'}
      />
      <Text padding={{ left: 'xlarge' }} className={css.checkboxText}>
        {getString('branchProtection.limitMergeStrategiesText')}
      </Text>
      {limitMergeStrats && (
        <Container padding={{ left: 'xlarge', top: 'large' }}>
          <Container padding={'small'} className={cx(css.widthContainer, css.greyContainer)}>
            <FormInput.CheckBox className={css.minText} label={getString('mergeCommit')} name={'mergeCommit'} />
            <FormInput.CheckBox className={css.minText} label={getString('squashMerge')} name={'squashMerge'} />
            <FormInput.CheckBox className={css.minText} label={getString('rebaseMerge')} name={'rebaseMerge'} />
          </Container>
        </Container>
      )}
      <hr className={css.dividerContainer} />
      <FormInput.CheckBox
        className={css.checkboxLabel}
        label={getString('branchProtection.autoDeleteTitle')}
        name={'autoDelete'}
      />
      <Text padding={{ left: 'xlarge' }} className={css.checkboxText}>
        {getString('branchProtection.autoDeleteText')}
      </Text>
    </Container>
  )
}

export default ProtectionRulesForm
