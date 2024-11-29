export function formatPostId(id: string): string {
  if (id.length < 8) return id;
  const prefix = id.slice(0, 4);
  const suffix = id.slice(-4);
  return `${prefix}...${suffix}`;
}
