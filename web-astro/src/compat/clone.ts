type Cloneable =
  | null
  | undefined
  | string
  | number
  | boolean
  | bigint
  | symbol
  | Date
  | RegExp
  | Map<unknown, unknown>
  | Set<unknown>
  | Cloneable[]
  | Record<string, unknown>;

function isPlainObject(value: unknown): value is Record<string, unknown> {
  if (value === null || typeof value !== "object") {
    return false;
  }

  const prototype = Object.getPrototypeOf(value);
  return prototype === Object.prototype || prototype === null;
}

function cloneRegExp(value: RegExp) {
  const cloned = new RegExp(value.source, value.flags);
  cloned.lastIndex = value.lastIndex;
  return cloned;
}

export function clone<T>(value: T, seen = new WeakMap<object, unknown>()): T {
  if (value === null || typeof value !== "object") {
    return value;
  }

  if (seen.has(value)) {
    return seen.get(value) as T;
  }

  if (value instanceof Date) {
    return new Date(value.getTime()) as T;
  }

  if (value instanceof RegExp) {
    return cloneRegExp(value) as T;
  }

  if (value instanceof Map) {
    const cloned = new Map();
    seen.set(value, cloned);
    for (const [key, entryValue] of value) {
      cloned.set(clone(key, seen), clone(entryValue, seen));
    }
    return cloned as T;
  }

  if (value instanceof Set) {
    const cloned = new Set();
    seen.set(value, cloned);
    for (const entry of value) {
      cloned.add(clone(entry, seen));
    }
    return cloned as T;
  }

  if (Array.isArray(value)) {
    const cloned: unknown[] = [];
    seen.set(value, cloned);
    for (const entry of value) {
      cloned.push(clone(entry, seen));
    }
    return cloned as T;
  }

  if (isPlainObject(value)) {
    const cloned: Record<string, unknown> = {};
    seen.set(value, cloned);
    for (const [key, entryValue] of Object.entries(value)) {
      cloned[key] = clone(entryValue, seen);
    }
    return cloned as T;
  }

  return value;
}

export default clone;
