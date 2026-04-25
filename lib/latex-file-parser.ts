export interface ParseTexFileResult {
  latex: string;
  problemCount: number;
  suggestedSource: string;
  warnings: string[];
}

const SKIP_SECTION_KEYWORDS = ['答案', '解析', '解答', '参考', '简析'];
const PROBLEM_MIN_CHARS = 60;

export function parseTexFile(content: string): ParseTexFileResult {
  const warnings: string[] = [];

  const suggestedSource = extractSuggestedSource(content);

  const docStart = content.indexOf('\\begin{document}');
  const body = docStart >= 0 ? content.slice(docStart + '\\begin{document}'.length) : content;

  const problems = extractProblems(body, warnings);

  if (problems.length === 0) {
    warnings.push('未检测到题目，请确认文件格式或手动粘贴内容。');
  }

  const latex = problems.map((p) => `\\begin{problem}\n${p}\n\\end{problem}`).join('\n\n');

  return { latex, problemCount: problems.length, suggestedSource, warnings };
}

function extractSuggestedSource(content: string): string {
  const m = content.match(/\\fancyhead\[R\]\{([^}]+)\}/);
  if (!m) return '';
  return m[1].replace(/\\,\s*/g, ' ').trim();
}

interface Section {
  title: string;
  content: string;
}

function splitBySections(body: string): Section[] {
  const sections: Section[] = [];
  const re = /\\section\{([^}]*)\}/g;
  let lastTitle = '';
  let lastEnd = 0;
  let isFirst = true;
  let m: RegExpExecArray | null;

  while ((m = re.exec(body)) !== null) {
    if (isFirst) {
      if (m.index > 0) sections.push({ title: '', content: body.slice(0, m.index) });
      isFirst = false;
    } else {
      sections.push({ title: lastTitle, content: body.slice(lastEnd, m.index) });
    }
    lastTitle = m[1];
    lastEnd = m.index + m[0].length;
  }

  sections.push({ title: lastTitle, content: body.slice(lastEnd) });
  return sections;
}

function isSkipSection(title: string): boolean {
  return SKIP_SECTION_KEYWORDS.some((kw) => title.includes(kw));
}

function findEnumerateBlocks(content: string): string[] {
  const blocks: string[] = [];
  let pos = 0;

  while (pos < content.length) {
    const startIdx = content.indexOf('\\begin{enumerate}', pos);
    if (startIdx < 0) break;

    let searchPos = startIdx + '\\begin{enumerate}'.length;
    // skip optional [...]
    if (content[searchPos] === '[') {
      const close = content.indexOf(']', searchPos);
      if (close >= 0) searchPos = close + 1;
    }
    const innerStart = searchPos;

    let depth = 1;
    while (searchPos < content.length && depth > 0) {
      if (content.startsWith('\\begin{enumerate}', searchPos)) {
        depth++;
        searchPos += '\\begin{enumerate}'.length;
      } else if (content.startsWith('\\end{enumerate}', searchPos)) {
        depth--;
        if (depth === 0) break;
        searchPos += '\\end{enumerate}'.length;
      } else {
        searchPos++;
      }
    }

    if (depth === 0) {
      blocks.push(content.slice(innerStart, searchPos));
      pos = searchPos + '\\end{enumerate}'.length;
    } else {
      break;
    }
  }

  return blocks;
}

function extractTopLevelItems(enumContent: string): string[] {
  const items: string[] = [];
  let depth = 0;
  let current = '';
  let pos = 0;

  while (pos < enumContent.length) {
    if (enumContent.startsWith('\\begin{', pos)) {
      depth++;
      const close = enumContent.indexOf('}', pos + 7);
      if (close < 0) {
        current += enumContent[pos++];
        continue;
      }
      current += enumContent.slice(pos, close + 1);
      pos = close + 1;
    } else if (enumContent.startsWith('\\end{', pos)) {
      depth--;
      const close = enumContent.indexOf('}', pos + 5);
      if (close < 0) {
        current += enumContent[pos++];
        continue;
      }
      current += enumContent.slice(pos, close + 1);
      pos = close + 1;
    } else if (depth === 0 && enumContent.startsWith('\\item', pos)) {
      // guard against \itemsep, \itemize, etc.
      const next = enumContent[pos + 5];
      const isItem = !next || next === ' ' || next === '\n' || next === '\t' || next === '[';
      if (isItem) {
        if (current.trim()) items.push(current.trim());
        current = '';
        pos += 5;
        if (enumContent[pos] === '[') {
          const close = enumContent.indexOf(']', pos);
          if (close >= 0) pos = close + 1;
        }
        continue;
      }
      current += enumContent[pos++];
    } else {
      current += enumContent[pos++];
    }
  }

  if (current.trim()) items.push(current.trim());
  return items;
}

function isProblemItem(item: string): boolean {
  // Skip theorem/remark notes starting with a bold label.
  // Handles both \textbf{label：} (colon inside) and \textbf{label}： (colon outside)
  if (/^\s*\\textbf\{[^}]*[：:][^}]*\}/.test(item)) return false;
  if (/^\s*\\textbf\{[^}]+\}[：:]/.test(item)) return false;

  // Multiple choice with tasks environment
  if (item.includes('\\begin{tasks}')) return true;
  // Inline A. B. C. D. options
  if (/\\item\s*\[A[\.\s]/.test(item)) return true;
  if (/A\.\s*(\\quad|\s)+B\./.test(item)) return true;
  // Minipage-based ABCD options
  if (/\\begin\{minipage\}/.test(item) && /[AB]\.\s/.test(item)) return true;
  // Fill-in-blank markers
  if (item.includes('\\underline')) return true;
  // Chinese exam blank placeholder （\quad）or (\quad)
  if (/（\\quad[）\s]|\(\\quad[）\s)]/.test(item)) return true;
  // Problem with clear question verbs and sufficient length,
  // but not if the item starts with \textbf (definition/theorem format)
  const stripped = item.replace(/%[^\n]*/g, '').replace(/\s+/g, ' ').trim();
  const startsWithBold = item.trimStart().startsWith('\\textbf');
  if (!startsWithBold && stripped.length >= PROBLEM_MIN_CHARS && /求|证明|解方程|计算|化简/.test(stripped)) return true;
  return false;
}

function extractProblems(body: string, warnings: string[]): string[] {
  const problems: string[] = [];
  const sections = splitBySections(body);

  for (const section of sections) {
    if (isSkipSection(section.title)) continue;
    for (const block of findEnumerateBlocks(section.content)) {
      for (const item of extractTopLevelItems(block)) {
        if (isProblemItem(item)) problems.push(item);
      }
    }
  }

  if (problems.length === 0 && warnings.length === 0) {
    warnings.push('未从 enumerate 环境中检测到题目。');
  }

  return problems;
}
