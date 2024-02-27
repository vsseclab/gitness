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
import { Layout, Text } from '@harnessio/uicore'
import { BookmarkBook } from 'iconoir-react'

import { FontVariation } from '@harnessio/design-system'
import { RepoPublicLabel } from 'components/RepoPublicLabel/RepoPublicLabel'
import type { GitInfoProps } from 'utils/GitUtils'
import { RepositoryPageHeader } from 'components/RepositoryPageHeader/RepositoryPageHeader'
import type { TypesRepository } from 'services/code'
import css from './RepositoryHeader.module.scss'
interface RepositoryHeaderProps extends Pick<GitInfoProps, 'repoMetadata'> {
  repoMetadata: TypesRepository
  className?: string
  isFile: boolean
}

export function RepositoryHeader(props: RepositoryHeaderProps) {
  const { repoMetadata, className, isFile } = props
  return (
    <RepositoryPageHeader
      className={isFile ? className : undefined}
      repoMetadata={repoMetadata}
      title={
        <Layout.Horizontal spacing="small" className={css.name}>
          <span className={css.customIcon}>
            <BookmarkBook />
          </span>
          {/* <Icon name={CodeIcon.Repo} size={20} /> */}
          <Text inline className={css.repoDropdown} font={{ variation: FontVariation.H4 }}>
            {repoMetadata.uid}
          </Text>
          <RepoPublicLabel isPublic={repoMetadata.is_public} />
        </Layout.Horizontal>
      }
      dataTooltipId="repositoryTitle"
    />
  )
}
