export const STANDARD_GRADES = [
  '初一',
  '初二',
  '初三',
  '高一',
  '高二',
  '高三',
] as const;

export type StandardGrade = typeof STANDARD_GRADES[number];

export function buildGradeOptions(grades?: readonly string[]) {
  const ordered: string[] = [...STANDARD_GRADES];
  const seen = new Set<string>(ordered);

  for (const grade of grades ?? []) {
    if (!grade || seen.has(grade)) {
      continue;
    }
    ordered.push(grade);
    seen.add(grade);
  }

  return ordered;
}
