export const STANDARD_GRADES = [
  '初一',
  '初二',
  '初三',
  '高一',
  '高二',
  '高三',
] as const;

export type StandardGrade = typeof STANDARD_GRADES[number];
