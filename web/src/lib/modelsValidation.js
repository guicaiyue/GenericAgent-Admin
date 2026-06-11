const IDENTIFIER_RE = /^[A-Za-z_][A-Za-z0-9_]*$/

const text = (value) => String(value ?? '').trim()
const numberValue = (value) => Number(value)

export function validateModelProfiles(profiles = []) {
  const counts = new Map()
  for (const profile of profiles || []) {
    const name = text(profile?.var_name)
    if (name) counts.set(name, (counts.get(name) || 0) + 1)
  }

  return (profiles || []).map((profile = {}, idx) => {
    const errors = []
    const warnings = []
    const varName = text(profile.var_name)
    const model = text(profile.model)
    const apiBase = text(profile.apibase)
    const maxRetries = numberValue(profile.max_retries ?? 3)
    const readTimeout = numberValue(profile.read_timeout ?? 300)

    if (!varName) errors.push('varNameRequired')
    else if (!IDENTIFIER_RE.test(varName)) errors.push('varNameInvalid')
    else if ((counts.get(varName) || 0) > 1) errors.push('varNameDuplicate')

    if (!model) errors.push('modelRequired')
    if (apiBase && !/^https?:\/\//i.test(apiBase)) warnings.push('apiBaseProtocol')
    if (!Number.isFinite(maxRetries) || maxRetries < 0) errors.push('maxRetriesInvalid')
    if (!Number.isFinite(readTimeout) || readTimeout <= 0) errors.push('readTimeoutInvalid')
    if (!text(profile.apikey)) warnings.push('apiKeyEmpty')

    return { idx, errors, warnings, ok: errors.length === 0 }
  })
}

export function modelValidationSummary(results = []) {
  const errors = results.reduce((sum, item) => sum + (item.errors?.length || 0), 0)
  const warnings = results.reduce((sum, item) => sum + (item.warnings?.length || 0), 0)
  return { errors, warnings, ready: results.length > 0 && errors === 0 }
}
