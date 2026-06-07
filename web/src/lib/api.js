const JSON_CONTENT_TYPE = 'application/json'
const DANGEROUS_CONFIRM_HEADER = 'X-GA-Confirm'
const DANGEROUS_CONFIRM_VALUE = 'dangerous'
const DEFAULT_TIMEOUT_MS = 15000

const isFormBody = (body) => typeof FormData !== 'undefined' && body instanceof FormData

export class ApiError extends Error {
  constructor({ message, endpoint = '', status = 0, timeout = false, cause = null } = {}) {
    super(message)
    this.name = 'ApiError'
    this.endpoint = endpoint
    this.status = status
    this.timeout = timeout
    this.cause = cause
  }
}

export const apiHeaders = ({ dangerous = false, headers = {}, body } = {}) => {
  const normalized = { ...(dangerous ? { [DANGEROUS_CONFIRM_HEADER]: DANGEROUS_CONFIRM_VALUE } : {}), ...headers }
  if (!isFormBody(body) && !Object.keys(normalized).some(k => k.toLowerCase() === 'content-type')) {
    normalized['Content-Type'] = JSON_CONTENT_TYPE
  }
  return normalized
}

export const parseApiResponse = async (res, url = '') => {
  const text = await res.text()
  let body = null
  if (text) {
    try { body = JSON.parse(text) }
    catch {
      if (!res.ok) throw new ApiError({ message: text.slice(0, 200) || `${res.status} ${res.statusText}`, endpoint: url, status: res.status })
      throw new ApiError({ message: `Expected JSON from ${url}, got ${text.slice(0, 40)}`, endpoint: url, status: res.status })
    }
  }
  if (!res.ok) throw new ApiError({ message: body?.detail || body?.error || text || `${res.status} ${res.statusText}`, endpoint: url, status: res.status })
  return body
}

const wrapFetch = async (url, req, timeoutMs) => {
  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), timeoutMs)
  try {
    return await fetch(url, { ...req, signal: controller.signal })
  } catch (err) {
    if (err.name === 'AbortError') {
      throw new ApiError({ message: `请求超时（${timeoutMs}ms）: ${url}`, endpoint: url, timeout: true, cause: err })
    }
    throw new ApiError({ message: err.message || '网络请求失败', endpoint: url, cause: err })
  } finally {
    clearTimeout(timeout)
  }
}

export const api = async (url, options = {}) => {
  const { dangerous = false, headers = {}, timeout = DEFAULT_TIMEOUT_MS, ...rest } = options
  const req = { ...rest, headers: apiHeaders({ dangerous, headers, body: rest.body }) }
  const res = await wrapFetch(url, req, timeout)
  if (!res.ok) await parseApiResponse(res, url)
  return res.json()
}

export const apiStream = async (url, options = {}) => {
  const { dangerous = false, headers = {}, timeout = DEFAULT_TIMEOUT_MS, ...rest } = options
  const req = { ...rest, headers: apiHeaders({ dangerous, headers, body: rest.body }) }
  const res = await wrapFetch(url, req, timeout)
  if (!res.ok) await parseApiResponse(res, url)
  return res
}
