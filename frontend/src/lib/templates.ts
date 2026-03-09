/**
 * Extract {{variable_name}} placeholders from a template string.
 * Returns unique variable names in order of first occurrence.
 */
export function extractVariables(template: string): string[] {
  const regex = /\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}/g
  const seen = new Set<string>()
  const order: string[] = []
  let m: RegExpExecArray | null
  while ((m = regex.exec(template)) !== null) {
    const name = m[1]
    if (!seen.has(name)) {
      seen.add(name)
      order.push(name)
    }
  }
  return order
}

import type { InputField } from './api-types'

/**
 * Convert variable names to InputField schema with default type "text".
 */
export function variablesToInputSchema(
  variables: string[],
  existing?: InputField[]
): InputField[] {
  const byKey = new Map<string, InputField>()
  for (const s of existing ?? []) {
    byKey.set(s.key, { key: s.key, type: s.type || 'text', label: s.label || s.key })
  }
  for (const key of variables) {
    if (!byKey.has(key)) {
      byKey.set(key, {
        key,
        type: 'text',
        label: key.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase()),
      })
    }
  }
  return Array.from(byKey.values())
}

/**
 * Render template with given inputs. Missing variables are left as {{var}}.
 */
export function renderTemplate(template: string, inputs: Record<string, unknown>): string {
  return template.replace(/\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}/g, (_, key) => {
    const v = inputs[key]
    return v != null ? String(v) : `{{${key}}}`
  })
}
