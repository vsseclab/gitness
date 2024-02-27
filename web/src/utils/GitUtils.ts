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

// Code copied from https://github.com/vweevers/is-git-ref-name-valid and
// https://github.com/vweevers/is-git-branch-name-valid (MIT, © Vincent Weevers)
// Last updated for git 2.29.0.

import type { IconName } from '@harnessio/icons'
import type {
  EnumWebhookTrigger,
  OpenapiContentInfo,
  OpenapiDirContent,
  OpenapiGetContentOutput,
  TypesCommit,
  TypesPullReq,
  TypesRepository
} from 'services/code'
import { getConfig } from 'services/config'
import { getErrorMessage } from './Utils'

export interface GitInfoProps {
  repoMetadata: TypesRepository
  gitRef: string
  resourcePath: string
  resourceContent: OpenapiGetContentOutput
  commitRef: string
  commits: TypesCommit[]
  pullReqMetadata: TypesPullReq
}
export interface RepoFormData {
  name: string
  description: string
  license: string
  defaultBranch: string
  gitignore: string
  addReadme: boolean
  isPublic: RepoVisibility
}
export interface ImportFormData {
  gitProvider: GitProviders
  hostUrl: string
  org: string
  repo: string
  username: string
  password: string
  name: string
  description: string
}

export interface ExportFormData {
  accountId: string
  token: string
  organization: string
  name: string
}

export interface ExportFormDataExtended extends ExportFormData {
  repoCount: number
}

export interface ImportSpaceFormData {
  gitProvider: GitProviders
  username: string
  password: string
  name: string
  description: string
  organization: string
  host: string
  importPipelineLabel: boolean
}

export enum RepoVisibility {
  PUBLIC = 'public',
  PRIVATE = 'private'
}

export enum RepoCreationType {
  IMPORT = 'import',
  CREATE = 'create',
  IMPORT_MULTIPLE = 'import_multiple'
}

export enum SpaceCreationType {
  IMPORT = 'import',
  CREATE = 'create'
}

export enum GitContentType {
  FILE = 'file',
  DIR = 'dir',
  SYMLINK = 'symlink',
  SUBMODULE = 'submodule'
}
export enum SettingsTab {
  webhooks = 'webhook',
  general = '/',
  branchProtection = 'rules'
}

export enum GitBranchType {
  ACTIVE = 'active',
  INACTIVE = 'inactive',
  YOURS = 'yours',
  ALL = 'all'
}

export enum GitRefType {
  BRANCH = 'branch',
  TAG = 'tag'
}

export enum PrincipalUserType {
  USER = 'user',
  SERVICE = 'service'
}

export enum SettingTypeMode {
  EDIT = 'edit',
  NEW = 'new'
}

export enum BranchTargetType {
  INCLUDE = 'include',
  EXCLUDE = 'exclude'
}

export interface BranchTargetOption {
  type: BranchTargetType
  title: string
}

export const branchTargetOptions: BranchTargetOption[] = [
  {
    type: BranchTargetType.INCLUDE,
    title: 'Include'
  },
  {
    type: BranchTargetType.EXCLUDE,
    title: 'Exclude'
  }
]

export enum GitCommitAction {
  DELETE = 'DELETE',
  CREATE = 'CREATE',
  UPDATE = 'UPDATE',
  MOVE = 'MOVE'
}

export enum PullRequestState {
  OPEN = 'open',
  MERGED = 'merged',
  CLOSED = 'closed'
}

export enum GitProviders {
  GITHUB = 'GitHub',
  GITHUB_ENTERPRISE = 'GitHub Enterprise',
  GITLAB = 'GitLab',
  GITLAB_SELF_HOSTED = 'GitLab Self-Hosted',
  BITBUCKET = 'Bitbucket',
  BITBUCKET_SERVER = 'Bitbucket Server',
  GITEA = 'Gitea',
  GOGS = 'Gogs'
}

export enum ConvertPipelineLabel {
  CONVERT = 'convert',
  IGNORE = 'ignore'
}

export const PullRequestFilterOption = {
  ...PullRequestState,
  // REJECTED: 'rejected',
  DRAFT: 'draft',
  YOURS: 'yours',
  ALL: 'all'
}

export const CodeIcon = {
  Logo: 'code' as IconName,
  PullRequest: 'git-pull' as IconName,
  Merged: 'code-merged' as IconName,
  Draft: 'code-draft' as IconName,
  PullRequestRejected: 'main-close' as IconName,
  Add: 'plus' as IconName,
  BranchSmall: 'code-branch-small' as IconName,
  Branch: 'code-branch' as IconName,
  Tag: 'main-tags' as IconName,
  Clone: 'code-clone' as IconName,
  Close: 'code-close' as IconName,
  CommitLight: 'code-commit-light' as IconName,
  CommitSmall: 'code-commit-small' as IconName,
  Commit: 'code-commit' as IconName,
  Copy: 'code-copy' as IconName,
  Delete: 'code-delete' as IconName,
  Edit: 'code-edit' as IconName,
  FileLight: 'code-file-light' as IconName,
  File: 'code-file' as IconName,
  Folder: 'code-folder' as IconName,
  History: 'code-history' as IconName,
  Info: 'code-info' as IconName,
  More: 'code-more' as IconName,
  Repo: 'code-repo' as IconName,
  Settings: 'code-settings' as IconName,
  Webhook: 'code-webhook' as IconName,
  InputSpinner: 'steps-spinne' as IconName,
  InputSearch: 'search' as IconName,
  Chat: 'code-chat' as IconName,
  Checks: 'main-tick' as IconName,
  ChecksSuccess: 'success-tick' as IconName
}

export const normalizeGitRef = (gitRef: string | undefined) => {
  if (isRefATag(gitRef)) {
    return gitRef
  } else if (isRefABranch(gitRef)) {
    return gitRef
  } else if (gitRef === '') {
    return ''
  } else if (gitRef && isGitRev(gitRef)) {
    return gitRef
  } else {
    return `refs/heads/${gitRef}`
  }
}

export const REFS_TAGS_PREFIX = 'refs/tags/'
export const REFS_BRANCH_PREFIX = 'refs/heads/'

export const FILE_VIEWED_OBSOLETE_SHA = 'ffffffffffffffffffffffffffffffffffffffff'

export function formatTriggers(triggers: EnumWebhookTrigger[]) {
  return triggers.map(trigger => {
    return trigger
      .split('_')
      .map(word => word.charAt(0).toUpperCase() + word.slice(1))
      .join(' ')
  })
}

export const handleUpload = (
  blob: File,
  setMarkdownContent: (data: string) => void,
  repoMetadata: TypesRepository | undefined,
  showError: (message: React.ReactNode, timeout?: number | undefined, key?: string | undefined) => void,
  standalone: boolean,
  routingId?: string
) => {
  const reader = new FileReader()
  // Set up a function to be called when the load event is triggered
  reader.onload = async function () {
    const markdown = await uploadImage(reader.result, showError, repoMetadata, standalone, routingId)
    setMarkdownContent(markdown) // Set the markdown content
  }
  reader.readAsArrayBuffer(blob) // This will trigger the onload function when the reading is complete
}

export const uploadImage = async (
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  fileBlob: any,
  showError: (message: React.ReactNode, timeout?: number | undefined, key?: string | undefined) => void,
  repoMetadata: TypesRepository | undefined,
  standalone: boolean,
  routingId?: string
) => {
  try {
    const response = await fetch(
      `${window.location.origin}${getConfig(
        `code/api/v1/repos/${repoMetadata?.path}/+/uploads${standalone || !routingId ? `` : `?routingId=${routingId}`}`
      )}`,
      {
        method: 'POST',
        headers: {
          Accept: 'application/json',
          'content-type': 'application/octet-stream'
        },
        body: fileBlob,
        redirect: 'follow'
      }
    )
    const result = await response.json()
    if (!response.ok && result) {
      showError(getErrorMessage(result))
      return ''
    }
    const filePath = result.file_path
    return `${window.location.origin}${getConfig(
      `code/api/v1/repos/${repoMetadata?.path}/+/uploads/${filePath}${
        standalone || !routingId ? `` : `?routingId=${routingId}`
      }`
    )}`
  } catch (exception) {
    showError(getErrorMessage(exception))
    return ''
  }
}

// eslint-disable-next-line no-control-regex
const BAD_GIT_REF_REGREX = /(^|[/.])([/.]|$)|^@$|@{|[\x00-\x20\x7f~^:?*[\\]|\.lock(\/|$)/
const BAD_GIT_BRANCH_REGREX = /^(-|HEAD$)/

function isGitRefValid(name: string, onelevel: boolean): boolean {
  return !BAD_GIT_REF_REGREX.test(name) && (!!onelevel || name.includes('/'))
}

export function isGitBranchNameValid(name: string): boolean {
  return isGitRefValid(name, true) && !BAD_GIT_BRANCH_REGREX.test(name)
}

export const isDir = (content: Nullable<OpenapiGetContentOutput>): boolean => content?.type === GitContentType.DIR
export const isFile = (content: Nullable<OpenapiGetContentOutput>): boolean => content?.type === GitContentType.FILE
export const isSymlink = (content: Nullable<OpenapiGetContentOutput>): boolean =>
  content?.type === GitContentType.SYMLINK
export const isSubmodule = (content: Nullable<OpenapiGetContentOutput>): boolean =>
  content?.type === GitContentType.SUBMODULE

export const findReadmeInfo = (content: Nullable<OpenapiGetContentOutput>): OpenapiContentInfo | undefined =>
  (content?.content as OpenapiDirContent)?.entries?.find(
    entry => entry.type === GitContentType.FILE && /^readme(.md)?$/.test(entry?.name?.toLowerCase() || '')
  )

export const findMarkdownInfo = (content: Nullable<OpenapiGetContentOutput>): OpenapiContentInfo | undefined =>
  content?.type === GitContentType.FILE && /.md$/.test(content?.name?.toLowerCase() || '') ? content : undefined

export const isRefATag = (gitRef: string | undefined) => gitRef?.includes(REFS_TAGS_PREFIX) || false
export const isRefABranch = (gitRef: string | undefined) => gitRef?.includes(REFS_BRANCH_PREFIX) || false

/**
 * Make a diff refs string to use in URL.
 * @param targetGitRef target git ref (base ref).
 * @param sourceGitRef source git ref (compare ref).
 * @returns A concatenation string of `targetGitRef...sourceGitRef`.
 */
export const makeDiffRefs = (targetGitRef: string, sourceGitRef: string) => `${targetGitRef}...${sourceGitRef}`

/**
 * Split a diff refs string into targetRef and sourceRef.
 * @param diffRefs diff refs string.
 * @returns An object of { targetGitRef, sourceGitRef }
 */
export const diffRefsToRefs = (diffRefs: string) => {
  const parts = diffRefs.split('...')

  return {
    targetGitRef: parts[0] || '',
    sourceGitRef: parts[1] || ''
  }
}

export const decodeGitContent = (content = '') => {
  try {
    // Decode base64 content for text file
    return decodeURIComponent(escape(window.atob(content)))
  } catch (_exception) {
    try {
      // Return original base64 content for binary file
      return content
    } catch (exception) {
      console.error(exception) // eslint-disable-line no-console
    }
  }
  return ''
}

// Check if gitRef is a git commit hash (https://github.com/diegohaz/is-git-rev, MIT © Diego Haz)
export const isGitRev = (gitRef = ''): boolean => /^[0-9a-f]{7,40}$/i.test(gitRef)

export const getProviderTypeMapping = (provider: GitProviders): string => {
  switch (provider) {
    case GitProviders.BITBUCKET_SERVER:
      return 'stash'
    case GitProviders.GITHUB_ENTERPRISE:
      return 'github'
    case GitProviders.GITLAB_SELF_HOSTED:
      return 'gitlab'
    default:
      return provider.toLowerCase()
  }
}

export const getOrgLabel = (gitProvider: string) => {
  switch (gitProvider) {
    case GitProviders.BITBUCKET:
      return 'importRepo.workspace'
    case GitProviders.BITBUCKET_SERVER:
      return 'importRepo.project'
    case GitProviders.GITLAB:
    case GitProviders.GITLAB_SELF_HOSTED:
      return 'importRepo.group'
    default:
      return 'importRepo.org'
  }
}

export const getOrgPlaceholder = (gitProvider: string) => {
  switch (gitProvider) {
    case GitProviders.BITBUCKET:
      return 'importRepo.workspacePlaceholder'
    case GitProviders.BITBUCKET_SERVER:
      return 'importRepo.projectPlaceholder'
    case GitProviders.GITLAB:
    case GitProviders.GITLAB_SELF_HOSTED:
      return 'importRepo.groupPlaceholder'
    default:
      return 'importRepo.orgPlaceholder'
  }
}

export const getProviders = () =>
  Object.values(GitProviders).map(value => {
    return { value, label: value }
  })
