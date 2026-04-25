const TEXT_COMMAND_PATTERN = /\\text\{([^{}]+)\}/g;
// Matches \item optionally followed by [label] — used in enumerate/tasks environments
const ITEM_PATTERN = /\\item(?:\[[^\]]*\])?\s*/g;

const MATH_DELIMITER_PATTERN = /\\\(|\\\[/;

export function normalizeLatexForDisplay(value: string): string {
  if (!value) {
    return value;
  }

  // Normalize \item commands before any math processing.
  const normalized = normalizeTextSegment(value);

  // If the content has no math delimiters, wrap the whole thing so MathJax renders it.
  if (!MATH_DELIMITER_PATTERN.test(normalized)) {
    return `\\(${normalizeMathSegment(normalized)}\\)`;
  }

  return rewriteLatexSegments(normalized, (s) => s, normalizeMathSegment);
}

function rewriteLatexSegments(
  value: string,
  textTransform: (segment: string) => string,
  mathTransform: (segment: string) => string
) {
  let cursor = 0;
  let output = '';

  while (cursor < value.length) {
    const next = findNextMathStart(value, cursor);
    if (!next) {
      output += textTransform(value.slice(cursor));
      break;
    }

    output += textTransform(value.slice(cursor, next.index));
    output += next.open;

    const end = value.indexOf(next.close, next.index + next.open.length);
    if (end === -1) {
      output += mathTransform(value.slice(next.index + next.open.length));
      break;
    }

    output += mathTransform(value.slice(next.index + next.open.length, end));
    output += next.close;
    cursor = end + next.close.length;
  }

  return output;
}

function findNextMathStart(value: string, fromIndex: number) {
  const candidates = [
    { open: '\\(', close: '\\)', index: value.indexOf('\\(', fromIndex) },
    { open: '\\[', close: '\\]', index: value.indexOf('\\[', fromIndex) },
  ].filter((item) => item.index >= 0);

  if (candidates.length === 0) {
    return null;
  }

  candidates.sort((left, right) => left.index - right.index);
  return candidates[0];
}

function normalizeTextSegment(segment: string) {
  // Replace \item sequences with A. B. C. D. labels
  let itemIndex = 0;
  const OPTION_LABELS = ['A', 'B', 'C', 'D', 'E', 'F'];
  const result = segment.replace(ITEM_PATTERN, () => {
    const label = OPTION_LABELS[itemIndex] ?? String(itemIndex + 1);
    itemIndex++;
    return `${label}. `;
  });
  return result.replace(TEXT_COMMAND_PATTERN, '$1');
}

function normalizeMathSegment(segment: string) {
  return segment.replace(/°/g, '^\\circ');
}
