<template>
  <div class="results-container">
    <!-- Status bar -->
    <div class="results-status">
      <div class="status-left">
        <span v-if="loading" aria-busy="true">Executing query...</span>
        <span v-else-if="error" class="status-error">✕ {{ error }}</span>
        <span v-else-if="result" class="status-ok">
          ✓ {{ result.message || `${result.rows ? result.rows.length : 0} row(s) returned` }}
          <template v-if="result.affected > 0">
            ({{ result.affected }} rows affected)
          </template>
        </span>
        <span v-else class="status-idle">Run a query to see results</span>
      </div>
      <div v-if="result && result.rows" class="status-right">
        <span class="row-count">{{ result.rows.length }} rows</span>
        <span v-if="result.columns" class="col-count">{{ result.columns.length }} columns</span>
      </div>
    </div>

    <!-- Results table -->
    <div v-if="result && result.columns && result.columns.length > 0" class="results-table-wrapper">
      <table class="results-table">
        <thead>
          <tr>
            <th class="row-num">#</th>
            <th
              v-for="col in result.columns"
              :key="col"
              :title="col"
            >
              {{ col }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, rowIdx) in result.rows" :key="rowIdx">
            <td class="row-num">{{ rowIdx + 1 }}</td>
            <td
              v-for="(cell, colIdx) in row"
              :key="colIdx"
              :class="cellClass(cell)"
              :title="formatCell(cell)"
            >
              {{ formatCell(cell) }}
            </td>
          </tr>
          <!-- Empty state: columns but no rows -->
          <tr v-if="result.rows.length === 0">
            <td :colspan="result.columns.length + 1" class="empty-result">
              No rows returned
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- No results placeholder -->
    <div v-else-if="!loading && !error && result" class="no-results">
      <div class="no-results-icon">📋</div>
      <p>{{ result.message || 'Query executed successfully' }}</p>
      <p v-if="result.affected > 0" class="affected-rows">
        {{ result.affected }} row(s) affected
      </p>
    </div>

    <!-- Empty state -->
    <div v-else-if="!loading && !error" class="no-results">
      <div class="no-results-icon">🔍</div>
      <p>Write a SQL query and click <strong>Run</strong> to see results</p>
      <p class="hint">Try: <code>SELECT * FROM table_name LIMIT 10</code></p>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ResultsTable',
  props: {
    result: { type: Object, default: null },
    error: { type: String, default: '' },
    loading: { type: Boolean, default: false },
  },
  methods: {
    formatCell(value) {
      if (value === null || value === undefined) return 'NULL'
      if (typeof value === 'object') return JSON.stringify(value)
      return String(value)
    },
    cellClass(value) {
      if (value === null || value === undefined) return 'cell-null'
      if (typeof value === 'number') return 'cell-number'
      return ''
    },
  },
}
</script>

<style scoped>
.results-container {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
  border-top: 1px solid var(--pico-muted-border-color);
}

.results-status {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.4rem 1rem;
  font-size: 0.8rem;
  background: var(--pico-card-background-color);
  border-bottom: 1px solid var(--pico-muted-border-color);
  flex-shrink: 0;
}

.status-left {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.status-error {
  color: var(--pico-del-color);
}

.status-ok {
  color: var(--pico-ins-color);
}

.status-idle {
  color: var(--pico-muted-color);
  font-style: italic;
}

.status-right {
  display: flex;
  gap: 0.75rem;
  color: var(--pico-muted-color);
}

.results-table-wrapper {
  flex: 1;
  overflow: auto;
}

.results-table {
  width: 100%;
  border-collapse: collapse;
  font-family: 'SF Mono', 'Fira Code', 'Fira Mono', 'Menlo', monospace;
  font-size: 0.78rem;
  margin: 0;
}

.results-table thead {
  position: sticky;
  top: 0;
  z-index: 10;
}

.results-table th {
  background: var(--pico-card-sectioning-background-color);
  padding: 0.5rem 0.75rem;
  text-align: left;
  font-weight: 600;
  border-bottom: 2px solid var(--pico-muted-border-color);
  white-space: nowrap;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.results-table td {
  padding: 0.35rem 0.75rem;
  border-bottom: 1px solid var(--pico-muted-border-color);
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.results-table tbody tr:hover {
  background: var(--pico-primary-background);
}

.results-table tbody tr:nth-child(even) {
  background: var(--pico-card-sectioning-background-color);
}

.results-table tbody tr:nth-child(even):hover {
  background: var(--pico-primary-background);
}

.row-num {
  color: var(--pico-muted-color);
  text-align: right;
  width: 3rem;
  min-width: 3rem;
  font-size: 0.7rem;
}

.cell-null {
  color: var(--pico-muted-color);
  font-style: italic;
}

.cell-number {
  text-align: right;
}

.empty-result {
  text-align: center;
  padding: 2rem !important;
  color: var(--pico-muted-color);
  font-style: italic;
}

.no-results {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: var(--pico-muted-color);
  text-align: center;
}

.no-results-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.no-results p {
  margin: 0.25rem 0;
}

.hint {
  font-size: 0.85rem;
  opacity: 0.7;
}

.hint code {
  background: var(--pico-card-sectioning-background-color);
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  font-size: 0.8rem;
}

.affected-rows {
  color: var(--pico-ins-color);
  font-weight: 600;
}
</style>
