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

import React, { useEffect, useMemo, useRef, useState } from 'react'
import { Container, useToaster } from '@harnessio/uicore'
import { LanguageDescription } from '@codemirror/language'
import { indentWithTab } from '@codemirror/commands'
import cx from 'classnames'
import type { ViewUpdate } from '@codemirror/view'
import type { Text } from '@codemirror/state'
import { languages } from '@codemirror/language-data'
import { markdown } from '@codemirror/lang-markdown'
import { java } from '@codemirror/lang-java'
import { EditorView, keymap, placeholder as placeholderExtension } from '@codemirror/view'
import { Compartment, EditorState, Extension } from '@codemirror/state'
import { color } from '@uiw/codemirror-extensions-color'
import { hyperLink } from '@uiw/codemirror-extensions-hyper-link'
import { githubLight, githubDark } from '@uiw/codemirror-themes-all'
import type { TypesRepository } from 'services/code'
import { useStrings } from 'framework/strings'
import { handleUpload } from 'utils/GitUtils'
import { handleFileDrop, handlePaste } from 'utils/Utils'
import css from './Editor.module.scss'

export interface EditorProps {
  content: string
  filename?: string
  forMarkdown?: boolean
  placeholder?: string
  readonly?: boolean
  autoFocus?: boolean
  className?: string
  extensions?: Extension
  maxHeight?: string | number
  viewRef?: React.MutableRefObject<EditorView | undefined>
  setDirty?: React.Dispatch<React.SetStateAction<boolean>>
  onChange?: (doc: Text, viewUpdate: ViewUpdate, isDirty: boolean) => void
  onViewUpdate?: (viewUpdate: ViewUpdate) => void
  darkTheme?: boolean
  repoMetadata: TypesRepository | undefined
  inGitBlame?: boolean
  standalone: boolean
  routingId?: string
}

export const Editor = React.memo(function CodeMirrorReactEditor({
  content,
  filename,
  forMarkdown,
  placeholder,
  readonly = false,
  autoFocus,
  className,
  extensions = new Compartment().of([]),
  maxHeight,
  viewRef,
  setDirty,
  onChange,
  onViewUpdate,
  darkTheme,
  repoMetadata,
  inGitBlame = false,
  standalone,
  routingId
}: EditorProps) {
  const { showError } = useToaster()
  const { getString } = useStrings()
  const view = useRef<EditorView>()
  const ref = useRef<HTMLDivElement>()
  const [fileData, setFile] = useState<File>()

  const languageConfig = useMemo(() => new Compartment(), [])
  const [markdownContent, setMarkdownContent] = useState('')
  const markdownLanguageSupport = useMemo(() => markdown({ codeLanguages: languages }), [])
  const style = useMemo(() => {
    if (maxHeight) {
      return {
        '--editor-max-height': Number.isInteger(maxHeight) ? `${maxHeight}px` : maxHeight
      } as React.CSSProperties
    }
  }, [maxHeight])
  const onChangeRef = useRef<EditorProps['onChange']>(onChange)

  useEffect(() => {
    onChangeRef.current = onChange
  }, [onChange, markdownContent])

  useEffect(() => {
    updateContentWithoutStateChange()
  }, [markdownContent]) // eslint-disable-line react-hooks/exhaustive-deps

  const updateContentWithoutStateChange = () => {
    setUploading(true)
    if (view.current && markdownContent && !inGitBlame) {
      const markdownInsert = fileData?.type.startsWith('image/') ? `![image](${markdownContent})` : `${markdownContent}`
      const range = view.current.state.selection.main
      const cursorPos = range.from
      const newCursorPos = cursorPos + markdownInsert.length
      // Create a transaction to update the document content
      const transaction = view.current.state.update({
        changes: {
          from: cursorPos,
          to: range.to,
          insert: markdownInsert
        },
        selection: { anchor: newCursorPos }
      })
      // Apply the transaction to update the view's state
      view.current.dispatch(transaction)
    }
  }

  const [uploading, setUploading] = useState(false)

  useEffect(() => {
    const editorView = new EditorView({
      doc: content,
      extensions: [
        extensions,

        color,
        hyperLink,
        darkTheme ? githubDark : githubLight,

        EditorView.lineWrapping,

        ...(placeholder ? [placeholderExtension(placeholder)] : []),

        keymap.of([indentWithTab]),

        ...(readonly ? [EditorState.readOnly.of(true), EditorView.editable.of(false)] : []),

        EditorView.updateListener.of(viewUpdate => {
          const isDirty = !cleanDoc.eq(viewUpdate.state.doc)
          setDirty?.(isDirty)
          onViewUpdate?.(viewUpdate)

          if (viewUpdate.docChanged) {
            onChangeRef.current?.(viewUpdate.state.doc, viewUpdate, isDirty)
          }
        }),

        /**
        languageConfig is a compartment that defaults to an empty array (no language support)
        at first, when a language is detected, languageConfig is used to reconfigure dynamically.
        @see https://codemirror.net/examples/config/
        */
        languageConfig.of([])
      ],
      parent: ref.current
    })

    view.current = editorView

    if (viewRef) {
      viewRef.current = editorView
    }

    const cleanDoc = editorView.state.doc

    if (autoFocus) {
      editorView.focus()
    }
    const messageElement = document.createElement('div')

    if (!inGitBlame) {
      // Create a new DOM element for the message
      messageElement.className = 'attachDiv'
      messageElement.textContent = uploading ? 'Uploading your files ...' : getString('attachText')
      editorView.dom.appendChild(messageElement)
    }

    return () => {
      messageElement.remove()
      editorView.destroy()
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  // Dynamically load language support based on filename. Note that
  // we need to configure languageSupport for Markdown separately to
  // enable syntax highlighting for all code blocks (multi-lang).
  useEffect(() => {
    if (forMarkdown) {
      view.current?.dispatch({ effects: languageConfig.reconfigure(markdownLanguageSupport) })
    } else if (filename) {
      LanguageDescription.matchFilename(languages, filename)
        ?.load()
        .then(languageSupport => {
          view.current?.dispatch({
            effects: languageConfig.reconfigure(
              languageSupport.language.name === 'markdown' ? markdownLanguageSupport : languageSupport
            )
          })
        })
    }
  }, [filename, forMarkdown, view, languageConfig, markdownLanguageSupport])

  useEffect(() => {
    if (filename) {
      let languageSupport
      if (
        filename.endsWith('.tf') ||
        filename.endsWith('.hcl') ||
        filename.endsWith('.tfstate') ||
        filename.endsWith('.tfvars')
      ) {
        languageSupport = java()
      }
      // Add other file extensions and their corresponding language modes
      if (languageSupport) {
        view.current?.dispatch({
          effects: languageConfig.reconfigure(languageSupport)
        })
      }
    }
  }, [filename, view, languageConfig, markdownLanguageSupport])
  const handleUploadCallback = (file: File) => {
    if (!inGitBlame) {
      setFile(file)
      handleUpload(file, setMarkdownContent, repoMetadata, showError, standalone, routingId)
    }
  }
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleDropForUpload = async (event: any) => {
    handleFileDrop(event, handleUploadCallback)
  }
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handlePasteForUpload = (event: { preventDefault: () => void; clipboardData: any }) => {
    handlePaste(event, handleUploadCallback)
  }

  return (
    <Container
      onDragOver={event => {
        event.preventDefault()
      }}
      onDrop={handleDropForUpload}
      onPaste={handlePasteForUpload}
      ref={ref}
      className={cx(css.editor, className, css.editorTest)}
      style={style}
    />
  )
})
