const FUZZY_TOLERANCE_RATIO = 0.4;

const SCORE_EXACT = 0;
const SCORE_PREFIX_BASE = 1000;
const SCORE_SUBSTRING_BASE = 2000;
const SCORE_FUZZY_BASE = 3000;

/**
 * Scores how well `name` matches `query`, lower is better. Returns null if
 * the two are too dissimilar to be considered a match at all.
 *
 * Matches are ranked in tiers - exact, then prefix, then substring, then
 * fuzzy (typo-tolerant) - so a small typo never outranks a real substring
 * hit, and fuzzy matches are only kept when the edit distance is small
 * relative to the query length (avoiding noisy, unrelated results).
 */
export function scoreMatch(name: string, query: string): number | null {
  const normalizedName = name.toLowerCase();
  const normalizedQuery = query.toLowerCase();

  if (normalizedName === normalizedQuery) {
    return SCORE_EXACT;
  }

  if (normalizedName.startsWith(normalizedQuery)) {
    return SCORE_PREFIX_BASE + (normalizedName.length - normalizedQuery.length);
  }

  const substringIndex = normalizedName.indexOf(normalizedQuery);
  if (substringIndex !== -1) {
    return SCORE_SUBSTRING_BASE + substringIndex;
  }

  const distance = levenshteinDistance(normalizedName, normalizedQuery);
  const tolerance = Math.max(1, Math.ceil(normalizedQuery.length * FUZZY_TOLERANCE_RATIO));
  if (distance <= tolerance) {
    return SCORE_FUZZY_BASE + distance;
  }

  return null;
}

export function levenshteinDistance(a: string, b: string): number {
  const matrix: number[][] = Array.from({ length: a.length + 1 }, () =>
    Array(b.length + 1).fill(0),
  );

  for (let i = 0; i <= a.length; i++) {
    matrix[i][0] = i;
  }
  for (let j = 0; j <= b.length; j++) {
    matrix[0][j] = j;
  }

  for (let i = 1; i <= a.length; i++) {
    for (let j = 1; j <= b.length; j++) {
      const cost = a[i - 1] === b[j - 1] ? 0 : 1;
      matrix[i][j] = Math.min(
        matrix[i - 1][j] + 1,
        matrix[i][j - 1] + 1,
        matrix[i - 1][j - 1] + cost,
      );
    }
  }

  return matrix[a.length][b.length];
}
