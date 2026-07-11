<template>
  <aside class="table-panel">
    <div class="panel-header">
      <h3>Tables</h3>
      <button
        class="outline contrast refresh-btn"
        @click="$emit('refresh')"
        :aria-busy="loading"
        title="Refresh tables"
      >
        🔄
      </button>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="panel-status">
      <span aria-busy="true">Loading tables...</span>
    </div>

    <!-- Empty state -->
    <div v-else-if="tables.length === 0" class="panel-status empty-state">
      <p>No tables found.</p>
      <p class="hint">Upload an Excel file to create tables.</p>
    </div>

    <!-- Table list -->
    <ul v-else class="table-list">
      <li
        v-for="table in tables"
        :key="table.name"
        class="table-item"
        :class="{ active: selectedTable === table.name }"
        @click="selectTable(table)"
        @dblclick="previewTable(table)"
      >
        <div class="table-name">
          <span class="table-icon">📋</span>
          <span class="table-label">{{ table.name }}</span>
          <span class="table-row-count">{{ table.row_count }}</span>
        </div>

        <!-- Schema detail (expandable) -->
        <div v-if="expandedTable === table.name" class="table-schema">
          <div
            v-for="col in table.columns"
            :key="col.cid"
            class="schema-column"
          >
            <span class="col-name">{{ col.name }}</span>
            <span class="col-type">{{ col.type }}</span>
            <span v-if="col.primary_key" class="col-pk">PK</span>
          </div>
        </div>
      </li>
    </ul>
  </aside>
</template>

<script>
export default {
  name: 'TablePanel',
  props: {
    tables: { type: Array, default: () => [] },
    selectedTable: { type: String, default: '' },
    loading: { type: Boolean, default: false },
  },
  emits: ['select-table', 'preview-table', 'refresh'],
  data() {
    return {
      expandedTable: null,
    }
  },
  methods: {
    selectTable(table) {
      this.expandedTable = this.expandedTable === table.name ? null : table.name
      this.$emit('select-table', table)
    },
    previewTable(table) {
      this.$emit('preview-table', table)
    },
  },
}
</script>

<style scoped>
.table-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-right: 1px solid var(--pico-muted-border-color);
  background: var(--pico-card-background-color);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--pico-muted-border-color);
}

.panel-header h3 {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--pico-muted-color);
}

.refresh-btn {
  padding: 0.2rem 0.4rem;
  font-size: 0.75rem;
  margin: 0;
}

.panel-status {
  padding: 1rem;
  text-align: center;
  color: var(--pico-muted-color);
  font-size: 0.875rem;
}

.empty-state p {
  margin: 0.25rem 0;
}

.hint {
  font-size: 0.8rem;
  opacity: 0.8;
}

.table-list {
  list-style: none;
  margin: 0;
  padding: 0;
  overflow-y: auto;
  flex: 1;
}

.table-item {
  border-bottom: 1px solid var(--pico-muted-border-color);
  cursor: pointer;
  transition: background 0.15s;
}

.table-item:hover {
  background: var(--pico-primary-background);
}

.table-item.active {
  background: var(--pico-primary);
  color: var(--pico-primary-inverse);
}

.table-name {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  font-size: 0.85rem;
}

.table-icon {
  font-size: 1rem;
  flex-shrink: 0;
}

.table-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: 'SF Mono', 'Fira Code', 'Fira Mono', monospace;
  font-size: 0.8rem;
}

.table-row-count {
  font-size: 0.75rem;
  opacity: 0.7;
  background: var(--pico-muted-border-color);
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  flex-shrink: 0;
}

.table-schema {
  padding: 0.25rem 1rem 0.5rem 2.5rem;
  font-size: 0.75rem;
  background: var(--pico-card-sectioning-background-color);
}

.schema-column {
  display: flex;
  gap: 0.5rem;
  padding: 0.1rem 0;
}

.col-name {
  font-family: 'SF Mono', 'Fira Code', 'Fira Mono', monospace;
  flex: 1;
}

.col-type {
  color: var(--pico-muted-color);
  font-style: italic;
}

.col-pk {
  color: var(--pico-primary);
  font-weight: bold;
  font-size: 0.65rem;
}
</style>
