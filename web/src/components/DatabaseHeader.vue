<template>
  <header class="app-header">
    <div class="header-content">
      <div class="brand">
        <span class="brand-icon">📊</span>
        <h1 class="brand-title">DBQuery</h1>
        <span class="brand-version">v{{ healthInfo.version }}</span>
        <span
          class="status-dot"
          :class="statusClass"
          :title="'Status: ' + healthInfo.status"
          @click="refreshHealth"
        ></span>
      </div>

      <div class="header-controls">
        <!-- Database selector -->
        <div class="db-selector">
          <label for="db-select" class="sr-only">Select Database</label>
          <select
            id="db-select"
            :value="modelValue"
            @change="onChange"
            :disabled="loading"
            class="db-dropdown"
          >
            <option value="" disabled>Select a database...</option>
            <option
              v-for="db in databases"
              :key="db.name"
              :value="db.name"
            >
              {{ db.name }}
              <template v-if="db.size > 0">({{ formatSize(db.size) }})</template>
            </option>
          </select>
        </div>

        <!-- Upload button -->
        <button
          class="upload-btn"
          @click="$emit('upload')"
          :aria-busy="uploading"
        >
          <span v-if="!uploading">📤 Upload Excel</span>
          <span v-else>Uploading...</span>
        </button>

        <!-- Refresh button -->
        <button
          class="refresh-btn outline contrast"
          @click="$emit('refresh')"
          title="Refresh databases"
        >
          🔄
        </button>

        <!-- User menu -->
        <div class="user-menu" v-if="username">
          <button
            class="user-btn"
            @click="toggleUserMenu"
            :aria-expanded="showUserMenu"
          >
            <span class="user-avatar">👤</span>
            <span class="user-name">{{ username }}</span>
          </button>

          <!-- Dropdown panel -->
          <div
            class="user-dropdown"
            v-if="showUserMenu"
          >
            <div class="dropdown-header">
              <span class="dropdown-username">{{ username }}</span>
            </div>
            <div class="dropdown-divider"></div>
            <button
              class="dropdown-item logout-item"
              @click="handleLogout"
            >
              🚪 Logout
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Click outside handler to close dropdown -->
    <div
      v-if="showUserMenu"
      class="dropdown-backdrop"
      @click="showUserMenu = false"
    ></div>

    <!-- Upload input (hidden) -->
    <input
      ref="fileInput"
      type="file"
      accept=".xlsx,.xlsm,.xltx,.xltm"
      style="display: none"
      @change="onFileSelected"
    />
  </header>
</template>

<script>
import { listDatabases, uploadExcel, healthCheck } from '../api.js'

export default {
  name: 'DatabaseHeader',
  props: {
    modelValue: { type: String, default: '' },
    loading: { type: Boolean, default: false },
    username: { type: String, default: '' },
  },
  emits: ['update:modelValue', 'upload', 'refresh', 'database-imported', 'logout'],
  data() {
    return {
      databases: [],
      uploading: false,
      healthInfo: { version: '?', status: 'unknown' },
      healthLoading: false,
      showUserMenu: false,
    }
  },
  computed: {
    statusClass() {
      if (this.healthLoading) return 'status-yellow'
      if (this.healthInfo.status === 'ok') return 'status-green'
      return 'status-red'
    },
  },
  async mounted() {
    await this.refreshHealth()
    await this.loadDatabases()
  },
  methods: {
    async refreshHealth() {
      this.healthLoading = true
      try {
        const result = await healthCheck()
        this.healthInfo = result
      } catch (err) {
        this.healthInfo = { version: this.healthInfo.version || '?', status: 'error' }
      } finally {
        this.healthLoading = false
      }
    },
    async loadDatabases() {
      try {
        const result = await listDatabases()
        this.databases = result.data || []
      } catch (err) {
        console.error('Failed to load databases:', err)
        this.databases = []
      }
    },
    onChange(event) {
      this.$emit('update:modelValue', event.target.value)
    },
    formatSize(bytes) {
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
      return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    },
    onFileSelected(event) {
      const file = event.target.files[0]
      if (!file) return
      this.uploadFile(file)
    },
    async uploadFile(file) {
      this.uploading = true
      try {
        const result = await uploadExcel(file)
        // Reload databases
        await this.loadDatabases()
        // Emit the database name to select it
        if (result.data && result.data.database) {
          this.$emit('update:modelValue', result.data.database)
        }
        this.$emit('database-imported', result)
      } catch (err) {
        alert('Upload failed: ' + err.message)
      } finally {
        this.uploading = false
        // Reset file input
        if (this.$refs.fileInput) {
          this.$refs.fileInput.value = ''
        }
      }
    },
    triggerUpload() {
      this.$refs.fileInput?.click()
    },
    toggleUserMenu() {
      this.showUserMenu = !this.showUserMenu
    },
    handleLogout() {
      this.showUserMenu = false
      this.$emit('logout')
    },
  },
}
</script>

<style scoped>
.app-header {
  background: var(--pico-card-background-color);
  border-bottom: 1px solid var(--pico-muted-border-color);
  padding: 0.5rem 1rem;
  height: var(--header-height, 64px);
  display: flex;
  align-items: center;
  position: sticky;
  top: 0;
  z-index: 100;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  gap: 1rem;
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}

.brand-icon {
  font-size: 1.5rem;
}

.brand-title {
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0;
  color: var(--pico-primary);
}

.brand-version {
  font-size: 0.7rem;
  font-weight: 400;
  color: var(--pico-muted-color);
  opacity: 0.7;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  cursor: pointer;
  flex-shrink: 0;
  transition: background-color 0.2s;
}

.status-green {
  background-color: #22c55e;
  box-shadow: 0 0 4px #22c55e66;
}

.status-red {
  background-color: #ef4444;
  box-shadow: 0 0 4px #ef444466;
}

.status-yellow {
  background-color: #eab308;
  box-shadow: 0 0 4px #eab30866;
}

.header-controls {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex: 1;
  justify-content: flex-end;
  max-width: 700px;
}

.db-selector {
  flex: 1;
  min-width: 200px;
  max-width: 350px;
}

.db-selector select {
  margin-bottom: 0;
  padding: 0.4rem 0.75rem;
  font-size: 0.875rem;
}

.db-dropdown {
  width: 100%;
}

.upload-btn {
  padding: 0.4rem 0.75rem;
  font-size: 0.875rem;
  margin-bottom: 0;
  white-space: nowrap;
}

.refresh-btn {
  padding: 0.4rem 0.6rem;
  font-size: 0.875rem;
  margin-bottom: 0;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0,0,0,0);
  white-space: nowrap;
  border: 0;
}

/* ─── User Menu ─────────────────────────────────────────── */

.user-menu {
  position: relative;
}

.user-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.35rem 0.65rem;
  font-size: 0.875rem;
  margin-bottom: 0;
  background: transparent;
  border: 1px solid var(--pico-muted-border-color);
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
  color: var(--pico-font-color);
}

.user-btn:hover {
  background: var(--pico-card-background-color);
  border-color: var(--pico-primary);
}

.user-avatar {
  font-size: 1.1rem;
}

.user-name {
  font-weight: 600;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-dropdown {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  min-width: 200px;
  background: var(--pico-card-background-color);
  border: 1px solid var(--pico-muted-border-color);
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  z-index: 200;
  overflow: hidden;
}

.dropdown-header {
  padding: 0.75rem 1rem;
}

.dropdown-username {
  font-weight: 700;
  font-size: 0.95rem;
  color: var(--pico-font-color);
}

.dropdown-divider {
  height: 1px;
  background: var(--pico-muted-border-color);
  margin: 0;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.65rem 1rem;
  font-size: 0.875rem;
  background: transparent;
  border: none;
  cursor: pointer;
  text-align: left;
  color: var(--pico-font-color);
  transition: background 0.12s;
}

.dropdown-item:hover {
  background: var(--pico-secondary-focus);
}

.logout-item {
  color: #dc2626;
}

.logout-item:hover {
  background: #fef2f2;
}

.dropdown-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 150;
}
</style>
