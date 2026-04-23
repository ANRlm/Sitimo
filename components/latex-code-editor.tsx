'use client';

import { useMemo } from 'react';
import CodeMirror, { type ReactCodeMirrorProps } from '@uiw/react-codemirror';
import { StreamLanguage } from '@codemirror/language';
import { stex } from '@codemirror/legacy-modes/mode/stex';
import { EditorView, placeholder as placeholderExtension } from '@codemirror/view';
import { cn } from '@/lib/utils';

const latexLanguage = StreamLanguage.define(stex);

type LatexCodeEditorProps = {
  value: string;
  onChange?: (value: string) => void;
  placeholder?: string;
  readOnly?: boolean;
  className?: string;
  minHeight?: number;
  basicSetup?: ReactCodeMirrorProps['basicSetup'];
};

export function LatexCodeEditor({
  value,
  onChange,
  placeholder,
  readOnly = false,
  className,
  minHeight = 220,
  basicSetup,
}: LatexCodeEditorProps) {
  const extensions = useMemo(
    () => [
      latexLanguage,
      EditorView.lineWrapping,
      EditorView.theme({
        '&': {
          height: '100%',
          backgroundColor: 'transparent',
          fontSize: '13px',
        },
        '.cm-scroller': {
          minHeight: `${minHeight}px`,
          fontFamily: 'var(--font-mono)',
        },
        '.cm-content': {
          padding: '14px 16px',
          caretColor: 'hsl(var(--primary))',
        },
        '.cm-gutters': {
          borderRight: '1px solid hsl(var(--border))',
          backgroundColor: 'hsl(var(--muted) / 0.3)',
          color: 'hsl(var(--muted-foreground))',
        },
        '.cm-activeLine': {
          backgroundColor: 'hsl(var(--primary) / 0.05)',
        },
        '.cm-activeLineGutter': {
          backgroundColor: 'hsl(var(--primary) / 0.08)',
        },
        '.cm-selectionBackground, &.cm-focused .cm-selectionBackground, ::selection': {
          backgroundColor: 'hsl(var(--accent) / 0.2) !important',
        },
        '&.cm-focused': {
          outline: 'none',
        },
      }),
      ...(placeholder ? [placeholderExtension(placeholder)] : []),
    ],
    [minHeight, placeholder]
  );

  return (
    <div className={cn('overflow-hidden rounded-xl border bg-background', className)}>
      <CodeMirror
        value={value}
        onChange={(nextValue) => onChange?.(nextValue)}
        extensions={extensions}
        basicSetup={
          basicSetup ?? {
            lineNumbers: true,
            foldGutter: true,
            highlightActiveLine: !readOnly,
            highlightActiveLineGutter: !readOnly,
            autocompletion: true,
            bracketMatching: true,
          }
        }
        editable={!readOnly}
        readOnly={readOnly}
      />
    </div>
  );
}
