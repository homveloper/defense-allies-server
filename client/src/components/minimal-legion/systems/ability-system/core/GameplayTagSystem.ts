export class GameplayTagSystem {
  private tags: Set<string> = new Set();
  private tagCounters: Map<string, number> = new Map();

  // Add a tag (with optional count for stacking)
  addTag(tag: string, count: number = 1): void {
    this.tags.add(tag);
    
    const currentCount = this.tagCounters.get(tag) || 0;
    this.tagCounters.set(tag, currentCount + count);
  }

  // Remove a tag (with optional count)
  removeTag(tag: string, count: number = 1): boolean {
    const currentCount = this.tagCounters.get(tag) || 0;
    
    if (currentCount <= 0) {
      return false; // Tag doesn't exist
    }

    const newCount = Math.max(0, currentCount - count);
    
    if (newCount === 0) {
      this.tags.delete(tag);
      this.tagCounters.delete(tag);
    } else {
      this.tagCounters.set(tag, newCount);
    }

    return true;
  }

  // Remove all instances of a tag
  removeAllTag(tag: string): boolean {
    if (this.tags.has(tag)) {
      this.tags.delete(tag);
      this.tagCounters.delete(tag);
      return true;
    }
    return false;
  }

  // Check if has exact tag
  hasTag(tag: string): boolean {
    return this.tags.has(tag);
  }

  // Get tag count (0 if not present)
  getTagCount(tag: string): number {
    return this.tagCounters.get(tag) || 0;
  }

  // Check if has any of the provided tags
  hasAnyTag(tags: string[]): boolean {
    return tags.some(tag => this.hasTag(tag));
  }

  // Check if has all of the provided tags
  hasAllTags(tags: string[]): boolean {
    return tags.every(tag => this.hasTag(tag));
  }

  // Hierarchical tag matching
  // Example: "Character.State.Stunned" matches "Character.State.*" or "Character.*"
  matchesPattern(pattern: string): boolean {
    // If pattern doesn't contain wildcards, do exact match
    if (!pattern.includes('*')) {
      return this.hasTag(pattern);
    }

    // Convert pattern to regex
    const regexPattern = pattern
      .replace(/\./g, '\\.')  // Escape dots
      .replace(/\*/g, '[^.]*'); // * matches any non-dot characters

    const regex = new RegExp(`^${regexPattern}$`);

    // Check if any tag matches the pattern
    for (const tag of this.tags) {
      if (regex.test(tag)) {
        return true;
      }
    }

    return false;
  }

  // Check if any tag matches any of the patterns
  matchesAnyPattern(patterns: string[]): boolean {
    return patterns.some(pattern => this.matchesPattern(pattern));
  }

  // Check if all patterns have matching tags
  matchesAllPatterns(patterns: string[]): boolean {
    return patterns.every(pattern => this.matchesPattern(pattern));
  }

  // Get all tags that match a pattern
  getTagsMatchingPattern(pattern: string): string[] {
    if (!pattern.includes('*')) {
      return this.hasTag(pattern) ? [pattern] : [];
    }

    const regexPattern = pattern
      .replace(/\./g, '\\.')
      .replace(/\*/g, '[^.]*');

    const regex = new RegExp(`^${regexPattern}$`);

    return Array.from(this.tags).filter(tag => regex.test(tag));
  }

  // Get all tags as array
  getAllTags(): string[] {
    return Array.from(this.tags).sort();
  }

  // Get all tags with their counts
  getAllTagsWithCounts(): Map<string, number> {
    return new Map(this.tagCounters);
  }

  // Clear all tags
  clear(): void {
    this.tags.clear();
    this.tagCounters.clear();
  }

  // Check if has no tags
  isEmpty(): boolean {
    return this.tags.size === 0;
  }

  // Get number of unique tags
  getTagCount(): number {
    return this.tags.size;
  }

  // Add multiple tags at once
  addTags(tags: string[]): void {
    tags.forEach(tag => this.addTag(tag));
  }

  // Remove multiple tags at once
  removeTags(tags: string[]): void {
    tags.forEach(tag => this.removeTag(tag));
  }

  // Create a snapshot for save/load
  getSnapshot(): { tags: string[], counts: [string, number][] } {
    return {
      tags: Array.from(this.tags),
      counts: Array.from(this.tagCounters.entries())
    };
  }

  // Restore from snapshot
  restoreFromSnapshot(snapshot: { tags: string[], counts: [string, number][] }): void {
    this.clear();
    this.tags = new Set(snapshot.tags);
    this.tagCounters = new Map(snapshot.counts);
  }

  // Debug helper
  toString(): string {
    const tagList = Array.from(this.tags).map(tag => {
      const count = this.tagCounters.get(tag) || 1;
      return count > 1 ? `${tag}(${count})` : tag;
    });
    
    return `Tags: [${tagList.join(', ')}]`;
  }

  // Static utility methods for tag validation and manipulation
  static isValidTag(tag: string): boolean {
    // Tags should contain only alphanumeric characters, dots, and underscores
    return /^[a-zA-Z0-9._]+$/.test(tag);
  }

  static normalizeTag(tag: string): string {
    // Remove extra spaces and convert to lowercase
    return tag.trim().toLowerCase();
  }

  static getTagParent(tag: string): string | null {
    const lastDotIndex = tag.lastIndexOf('.');
    return lastDotIndex > 0 ? tag.substring(0, lastDotIndex) : null;
  }

  static getTagChildren(parentTag: string, allTags: string[]): string[] {
    const pattern = `${parentTag}.`;
    return allTags.filter(tag => 
      tag.startsWith(pattern) && 
      tag.indexOf('.', pattern.length) === -1 // Direct children only
    );
  }
}