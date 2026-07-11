<template>
  <div class="app-layout">
    <!-- Login view when not authenticated -->
    <LoginView v-if="!authenticated" @auth-success="onAuthSuccess" />

    <!-- Main app when authenticated -->
    <template v-else>
      <!-- Header -->
      <DatabaseHeader
        v-model="currentDatabase"
        :loading="loadingDatabases"
        :uploading="false"
        :username="username"
        @upload="triggerUpload"
        @refresh="refreshAll"
        @database-imported="onDatabaseImported"
        @logout="onLogout"
        ref="headerRef"
      />

      <!-- Main content area -->
      <div class="main-content">
        <!-- Left panel: Tables -->
        <div class="left-panel" :class="{ collapsed: !currentDatabase }">
          <TablePanel
            :tables="tables"
            :selected-table="selectedTable"
            :loading="loadingTables"
            @select-table="onSelectTable"
            @preview-table="onPreviewTable"
            @refresh="loadTables"
          />
        </div>

        <!-- Right panel: Editor + Results -->
        <div class="right-panel">
          <!-- Editor section -->
          <div class="editor-section" :style="{ flex: editorFlex }">
            <SqlEditor
              :database="currentDatabase"
              :running="queryRunning"
              @run-query="onRunQuery"
              ref="editorRef"
            />
          </div>

          <!-- Resize handle -->
          <div
            class="resize-handle"
            @mousedown="startResize"
          ></div>

          <!-- Results section -->
          <div class="results-section" :style="{ flex: resultsFlex }">
            <ResultsTable
              :result="queryResult"
              :error="queryError"
              :loading="queryRunning"
            />
          </div>
        </div>
      </div>

      <!-- Hidden file input for upload -->
      <input
        ref="fileInput"
        type="file"
        accept=".xlsx,.xlsm,.xltx,.xltm"
        style="display: none"
        @change="onFileSelected"
      />
    </template>
  </div>
</template>

<script>
import LoginView from './components/Login.vue'
import DatabaseHeader from './components/DatabaseHeader.vue'
import TablePanel from './components/TablePanel.vue'
import SqlEditor from './components/SqlEditor.vue'
import ResultsTable from './components/ResultsTable.vue'
import { listDatabases, listTables, executeQuery, getTableData, uploadExcel, getMe, logout } from './api.js'

export default {
  name: 'App',
  components: {
    LoginView,
    DatabaseHeader,
    TablePanel,
    SqlEditor,
    ResultsTable,
  },
  data() {
    return {
      // Auth state
      authenticated: false,
      username: '',

      // Database state
      currentDatabase: '',
      databases: [],
      loadingDatabases: false,
      uploading: false,

      // Tables state
      tables: [],
      selectedTable: '',
      loadingTables: false,

      // Query state
      queryRunning: false,
      queryResult: null,
      queryError: '',

      // Layout state
      editorFlex: 1,
      resultsFlex: 1,
      isResizing: false,
    }
  },
  async mounted() {
    // Listen for auth expiry events from the API layer
    window.addEventListener('auth-expired', this.onAuthExpired)

    // Check if user has a stored token and validate it
    await this.checkAuth()
  },
  beforeUnmount() {
    window.removeEventListener('auth-expired', this.onAuthExpired)
  },
  methods: {
    async checkAuth() {
      const token = localStorage.getItem('auth_token')
      if (!token) {
        this.authenticated = false
        return
      }

      try {
        const result = await getMe()
        this.username = result.data?.username || ''
        this.authenticated = true
        await this.loadDatabases()
      } catch (err) {
        // Token is invalid or expired
        this.authenticated = false
        this.username = ''
        localStorage.removeItem('auth_token')
        localStorage.removeItem('auth_username')
      }
    },

    onAuthSuccess({ token, username }) {
      this.authenticated = true
      this.username = username
      this.loadDatabases()
    },

    onAuthExpired() {
      this.authenticated = false
      this.username = ''
      this.currentDatabase = ''
      this.databases = []
      this.tables = []
      this.queryResult = null
    },

    async onLogout() {
      try {
        await logout()
      } catch (err) {
        // Ignore logout errors — clear locally regardless
      }
      localStorage.removeItem('auth_token')
      localStorage.removeItem('auth_username')
      this.authenticated = false
      this.username = ''
      this.currentDatabase = ''
      this.databases = []
      this.tables = []
      this.queryResult = null
    },

    async loadDatabases() {
      this.loadingDatabases = true
      try {
        const result = await listDatabases()
        this.databases = result.data || []
        // Auto-select if only one database
        if (this.databases.length === 1 && !this.currentDatabase) {
          this.currentDatabase = this.databases[0].name
        }
      } catch (err) {
        console.error('Failed to load databases:', err)
      } finally {
        this.loadingDatabases = false
      }
    },

    async loadTables() {
      if (!this.currentDatabase) {
        this.tables = []
        return
      }
      this.loadingTables = true
      try {
        const result = await listTables(this.currentDatabase)
        this.tables = result.data || []
      } catch (err) {
        console.error('Failed to load tables:', err)
        this.tables = []
      } finally {
        this.loadingTables = false
      }
    },

    async refreshAll() {
      await this.loadDatabases()
      await this.loadTables()
    },

    onSelectTable(table) {
      this.selectedTable = table.name
    },

    onPreviewTable(table) {
      // Insert a SELECT query for this table into the editor
      const query = `SELECT * FROM [${table.name}] LIMIT 100`
      if (this.$refs.editorRef) {
        this.$refs.editorRef.query = query
      }
    },

    async onRunQuery(query) {
      if (!this.currentDatabase) {
        this.queryError = 'Please select a database first'
        return
      }

      this.queryRunning = true
      this.queryError = ''
      this.queryResult = null

      try {
        const result = await executeQuery(this.currentDatabase, query)
        this.queryResult = result.data
      } catch (err) {
        this.queryError = err.message
        this.queryResult = null
      } finally {
        this.queryRunning = false
      }
    },

    triggerUpload() {
      this.$refs.fileInput?.click()
    },

    async onFileSelected(event) {
      const file = event.target.files[0]
      if (!file) return

      this.uploading = true
      try {
        const result = await uploadExcel(file)
        await this.loadDatabases()
        if (result.data && result.data.database) {
          this.currentDatabase = result.data.database
        }
      } catch (err) {
        alert('Upload failed: ' + err.message)
      } finally {
        this.uploading = false
        if (this.$refs.fileInput) {
          this.$refs.fileInput.value = ''
        }
      }
    },

    onDatabaseImported(result) {
      // Refresh tables for the new database
      this.loadTables()
    },

    // Resize handler for editor/results split
    startResize(event) {
      this.isResizing = true
      const startY = event.clientY
      const startEditorFlex = this.editorFlex
      const startResultsFlex = this.resultsFlex
      const totalFlex = startEditorFlex + startResultsFlex
      const container = event.target.parentElement
      const containerHeight = container.clientHeight

      const onMouseMove = (e) => {
        if (!this.isResizing) return
        const deltaY = e.clientY - startY
        const deltaFlex = (deltaY / containerHeight) * totalFlex
        this.editorFlex = Math.max(0.2, startEditorFlex + deltaFlex)
        this.resultsFlex = Math.max(0.2, startResultsFlex - deltaFlex)
      }

      const onMouseUp = () => {
        this.isResizing = false
        document.removeEventListener('mousemove', onMouseMove)
        document.removeEventListener('mouseup', onMouseUp)
      }

      document.addEventListener('mousemove', onMouseMove)
      document.addEventListener('mouseup', onMouseUp)
    },
  },
  watch: {
    currentDatabase() {
      this.loadTables()
      this.selectedTable = ''
      this.queryResult = null
      this.queryError = ''
    },
  },
}
</script>

<style>
/* Global layout styles */
html, body {
  height: 100%;
  overflow: hidden;
}

#app {
  height: 100%;
}

.app-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.main-content {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.left-panel {
  width: var(--left-panel-width, 280px);
  min-width: 200px;
  max-width: 400px;
  flex-shrink: 0;
  overflow: hidden;
  border-right: 1px solid var(--pico-muted-border-color);
  transition: width 0.2s ease;
}

.left-panel.collapsed {
  width: 0;
  min-width: 0;
  border-right: none;
}

.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

.editor-section {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 100px;
}

.results-section {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 100px;
}

.resize-handle {
  height: 4px;
  background: var(--pico-muted-border-color);
  cursor: row-resize;
  flex-shrink: 0;
  transition: background 0.15s;
  position: relative;
}

.resize-handle:hover,
.resize-handle:active {
  background: var(--pico-primary);
}

/* PicoCSS overrides for better spacing */
.container-fluid {
  padding: 0;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .left-panel {
    width: 200px;
    min-width: 150px;
  }
}
</style>
