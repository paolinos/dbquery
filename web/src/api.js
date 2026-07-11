// HOST
const API_BASE = 'http://localhost:8080/api'

/**
 * Get the stored auth token from localStorage.
 */
function getAuthToken() {
  return localStorage.getItem('auth_token')
}

/**
 * Generic fetch wrapper with error handling.
 * Automatically attaches the Authorization header if a token is stored.
 */
async function request(url, options = {}) {
  const token = getAuthToken()

  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  }

  // Attach JWT token if available
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const config = {
    ...options,
    headers,
  }

  // Remove Content-Type for FormData
  if (options.body instanceof FormData) {
    delete config.headers['Content-Type']
  }

  const response = await fetch(`${API_BASE}${url}`, config)

  // If 401, clear stored auth and redirect
  if (response.status === 401) {
    localStorage.removeItem('auth_token')
    localStorage.removeItem('auth_username')
    // Dispatch a custom event so the app can react
    window.dispatchEvent(new CustomEvent('auth-expired'))
    throw new Error('Session expired. Please log in again.')
  }

  const data = await response.json()

  if (!response.ok) {
    throw new Error(data.details || data.error || 'Request failed')
  }

  return data
}

// ─── Auth API ──────────────────────────────────────────────

/**
 * Check if any users exist in the system.
 * @returns {Promise<{data: {has_users: boolean}}>}
 */
export function hasUsers() {
  return request('/auth/has-users')
}

/**
 * Login with username and password.
 * @param {string} username
 * @param {string} password
 * @returns {Promise<{data: {token: string, username: string}}>}
 */
export function login(username, password) {
  return request('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
}

/**
 * Register a new user (only works when no users exist).
 * @param {string} username
 * @param {string} password
 * @returns {Promise<{data: {token: string, username: string}}>}
 */
export function register(username, password) {
  return request('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
}

/**
 * Logout — revokes the current token server-side.
 * @returns {Promise<{message: string}>}
 */
export function logout() {
  return request('/auth/logout', { method: 'POST' })
}

/**
 * Get current user info from the server.
 * @returns {Promise<{data: {user_id: number, username: string}}>}
 */
export function getMe() {
  return request('/auth/me')
}

// ─── Database API ──────────────────────────────────────────

/**
 * Health check — returns server version, status, and timestamp.
 * @returns {Promise<{version: string, status: string, timestamp: string}>}
 */
export function healthCheck() {
  return request('/health')
}

/**
 * List all databases.
 * @returns {Promise<{data: Array, message: string}>}
 */
export function listDatabases() {
  return request('/databases')
}

/**
 * List all tables in a database.
 * @param {string} db - Database name
 * @returns {Promise<{data: Array, message: string}>}
 */
export function listTables(db) {
  return request(`/databases/${encodeURIComponent(db)}/tables`)
}

/**
 * Get table schema.
 * @param {string} db - Database name
 * @param {string} table - Table name
 * @returns {Promise<{data: Object, message: string}>}
 */
export function getTableSchema(db, table) {
  return request(`/databases/${encodeURIComponent(db)}/tables/${encodeURIComponent(table)}/schema`)
}

/**
 * Get paginated table data.
 * @param {string} db - Database name
 * @param {string} table - Table name
 * @param {number} page - Page number (1-based)
 * @param {number} perPage - Items per page
 * @returns {Promise<{data: Object, message: string}>}
 */
export function getTableData(db, table, page = 1, perPage = 100) {
  return request(`/databases/${encodeURIComponent(db)}/tables/${encodeURIComponent(table)}/data?page=${page}&per_page=${perPage}`)
}

/**
 * Execute a SQL query.
 * @param {string} db - Database name
 * @param {string} query - SQL query
 * @returns {Promise<{data: {columns: string[], rows: any[][]}, message: string}>}
 */
export function executeQuery(db, query) {
  return request(`/databases/${encodeURIComponent(db)}/query`, {
    method: 'POST',
    body: JSON.stringify({ query }),
  })
}

/**
 * Get autocomplete suggestions for a database.
 * @param {string} db - Database name
 * @returns {Promise<{data: Array, message: string}>}
 */
export function getAutocomplete(db) {
  return request(`/databases/${encodeURIComponent(db)}/autocomplete`)
}

/**
 * Upload an Excel file.
 * @param {File} file - Excel file to upload
 * @returns {Promise<{data: Object, message: string}>}
 */
export function uploadExcel(file) {
  const formData = new FormData()
  formData.append('file', file)
  return request('/upload', {
    method: 'POST',
    body: formData,
  })
}
