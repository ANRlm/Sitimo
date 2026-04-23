export function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / (1000 * 60));
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (diffMins < 60) {
    return `${Math.max(diffMins, 1)} 分钟前`;
  }

  if (diffHours < 24) {
    return `${diffHours} 小时前`;
  }

  if (diffDays < 7) {
    return `${diffDays} 天前`;
  }

  return formatAbsoluteDateTime(dateString);
}

export function formatAbsoluteDateTime(dateString: string): string {
  const date = new Date(dateString);
  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, '0');
  const day = `${date.getDate()}`.padStart(2, '0');
  const hours = `${date.getHours()}`.padStart(2, '0');
  const minutes = `${date.getMinutes()}`.padStart(2, '0');

  return `${year}-${month}-${day} ${hours}:${minutes}`;
}

export type HighlightPart = {
  text: string;
  highlighted: boolean;
};

function normalizeLatexPreview(source: string): string {
  return source
    .replace(/\\\(|\\\)|\\\[|\\\]/g, '')
    .replace(/\\(?:left|right)/g, '')
    .replace(/\\[a-zA-Z]+\*?(?:\[[^\]]*\])?\{([^{}]*)\}/g, '$1')
    .replace(/\\[a-zA-Z]+\*?/g, ' ')
    .replace(/[{}]/g, '')
    .replace(/\s+/g, ' ');
}

export function stripLatex(source: string): string {
  return normalizeLatexPreview(source).trim();
}

export function highlightText(text: string, query: string): HighlightPart[] {
  if (!query.trim()) {
    return [{ text, highlighted: false }];
  }

  const escaped = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const regex = new RegExp(`(${escaped})`, 'ig');
  const parts = text.split(regex).filter(Boolean);

  return parts.map((part) => ({
    text: part,
    highlighted: part.toLowerCase() === query.toLowerCase(),
  }));
}

function splitMarkedText(text: string): HighlightPart[] {
  const parts: HighlightPart[] = [];
  const regex = /<mark>([\s\S]*?)<\/mark>/gi;
  let lastIndex = 0;

  for (const match of text.matchAll(regex)) {
    const index = match.index ?? 0;
    const before = normalizeLatexPreview(text.slice(lastIndex, index));
    if (before.trim()) {
      parts.push({ text: before, highlighted: false });
    }

    const highlighted = normalizeLatexPreview(match[1] ?? '');
    if (highlighted.trim()) {
      parts.push({ text: highlighted, highlighted: true });
    }
    lastIndex = index + match[0].length;
  }

  const tail = normalizeLatexPreview(text.slice(lastIndex));
  if (tail.trim()) {
    parts.push({ text: tail, highlighted: false });
  }

  if (parts.length === 0) {
    return [{ text: stripLatex(text), highlighted: false }];
  }

  parts[0].text = parts[0].text.trimStart();
  parts[parts.length - 1].text = parts[parts.length - 1].text.trimEnd();
  return parts.filter((part) => part.text.length > 0);
}

export function buildSearchPreview(text: string, query: string): HighlightPart[] {
  if (text.includes('<mark>')) {
    return splitMarkedText(text);
  }
  return highlightText(stripLatex(text), query);
}
