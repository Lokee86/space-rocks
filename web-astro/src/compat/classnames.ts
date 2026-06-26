type ClassValue =
  | string
  | number
  | null
  | undefined
  | false
  | ClassValue[]
  | Record<string, unknown>;

function appendClassName(current: string, next: string) {
  return current ? `${current} ${next}` : next;
}

function resolveClassValue(value: ClassValue): string {
  if (!value) {
    return "";
  }

  if (typeof value === "string" || typeof value === "number") {
    return String(value);
  }

  if (Array.isArray(value)) {
    return classNames(...value);
  }

  let result = "";

  for (const [key, include] of Object.entries(value)) {
    if (include) {
      result = appendClassName(result, key);
    }
  }

  return result;
}

export function classNames(...values: ClassValue[]) {
  let result = "";

  for (const value of values) {
    const className = resolveClassValue(value);
    if (className) {
      result = appendClassName(result, className);
    }
  }

  return result;
}

export default classNames;
