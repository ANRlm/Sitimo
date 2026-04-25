const TEXT_COMMAND_PATTERN = /\\text\{([^{}]+)\}/g;
// Matches \item or \task optionally followed by [label]
const ITEM_PATTERN = /\\(?:item|task)(?:\[[^\]]*\])?\s*/g;

// Detects any math delimiter: \(, \[, $, $$
const HAS_MATH_PATTERN = /\\\(|\\\[|\$\$?/;

export function normalizeLatexForDisplay(value: string): string {
  if (!value) {
    return value;
  }

  // Pre-process \item/\task into A. B. C. D. labels across the whole string
  // (they always appear in text segments, never inside math)
  const preprocessed = replaceItemCommands(value);

  // If no math delimiters at all, wrap entire content for MathJax
  if (!HAS_MATH_PATTERN.test(preprocessed)) {
    return `\\(${normalizeMathSegment(preprocessed)}\\)`;
  }

  return rewriteLatexSegments(preprocessed);
}

/**
 * Replace \item / \task sequences with A. B. C. D. labels.
 * These commands always appear outside math delimiters.
 */
function replaceItemCommands(value: string): string {
  let itemIndex = 0;
  const OPTION_LABELS = ['A', 'B', 'C', 'D', 'E', 'F'];
  return value.replace(ITEM_PATTERN, () => {
    const label = OPTION_LABELS[itemIndex] ?? String(itemIndex + 1);
    itemIndex++;
    return `${label}. `;
  });
}

/**
 * Walk through the string, identify math vs text segments, and apply
 * appropriate transforms. Supports \(...\), \[...\], $...$, $$...$$
 */
function rewriteLatexSegments(value: string): string {
  let cursor = 0;
  let output = '';

  while (cursor < value.length) {
    const next = findNextMathStart(value, cursor);
    if (!next) {
      // Remaining text segment — apply \text{} cleanup
      output += normalizeTextSegment(value.slice(cursor));
      break;
    }

    // Text before math
    output += normalizeTextSegment(value.slice(cursor, next.index));
    output += next.open;

    const end = value.indexOf(next.close, next.index + next.open.length);
    if (end === -1) {
      // Unclosed math — treat rest as math
      output += normalizeMathSegment(value.slice(next.index + next.open.length));
      break;
    }

    output += normalizeMathSegment(value.slice(next.index + next.open.length, end));
    output += next.close;
    cursor = end + next.close.length;
  }

  return output;
}

type MathDelimiter = { open: string; close: string; index: number };

function findNextMathStart(value: string, fromIndex: number): MathDelimiter | null {
  const candidates: MathDelimiter[] = [
    { open: '\\(', close: '\\)', index: value.indexOf('\\(', fromIndex) },
    { open: '\\[', close: '\\]', index: value.indexOf('\\[', fromIndex) },
  ].filter((c) => c.index >= 0);

  // Check for $$ before $ to avoid treating $$ as two separate $
  const dblIdx = value.indexOf('$$', fromIndex);
  if (dblIdx >= 0) {
    candidates.push({ open: '$$', close: '$$', index: dblIdx });
  }

  // Single $ — only if not part of $$
  let singleIdx = value.indexOf('$', fromIndex);
  while (singleIdx >= 0) {
    if (value[singleIdx + 1] !== '$' && (singleIdx === 0 || value[singleIdx - 1] !== '$')) {
      candidates.push({ open: '$', close: '$', index: singleIdx });
      break;
    }
    singleIdx = value.indexOf('$', singleIdx + 1);
  }

  if (candidates.length === 0) return null;
  candidates.sort((a, b) => a.index - b.index);
  return candidates[0];
}

function normalizeTextSegment(segment: string): string {
  return segment.replace(TEXT_COMMAND_PATTERN, '$1');
}

function normalizeMathSegment(segment: string): string {
  return segment.replace(/°/g, '^\\circ');
}
