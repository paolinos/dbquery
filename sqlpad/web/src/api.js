// HOST
const API_BASE = 'http://localhost:8080/api'

/**
 * Generic fetch wrapper with error handling.
 */
async function request(url, options = {}) {
  const config = {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  }

  // Remove Content-Type for FormData
  if (options.body instanceof FormData) {
    delete config.headers['Content-Type']
  }

  const response = await fetch(`${API_BASE}${url}`, config)
  const data = await response.json()

  if (!response.ok) {
    throw new Error(data.details || data.error || 'Request failed')
  }

  return data
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
