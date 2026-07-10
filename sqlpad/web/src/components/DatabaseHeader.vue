<template>
  <header class="app-header">
    <div class="header-content">
      <div class="brand">
        <span class="brand-icon">📊</span>
        <h1 class="brand-title">SQLPad</h1>
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
      </div>
    </div>

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
import { listDatabases, uploadExcel } from '../api.js'

export default {
  name: 'DatabaseHeader',
  props: {
    modelValue: { type: String, default: '' },
    loading: { type: Boolean, default: false },
  },
  emits: ['update:modelValue', 'upload', 'refresh', 'database-imported'],
  data() {
    return {
      databases: [],
      uploading: false,
    }
  },
  async mounted() {
    await this.loadDatabases()
  },
  methods: {
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

.header-controls {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex: 1;
  justify-content: flex-end;
  max-width: 600px;
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
</style>
