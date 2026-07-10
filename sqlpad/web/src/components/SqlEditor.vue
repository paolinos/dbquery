<template>
  <div class="sql-editor-container">
    <!-- Toolbar -->
    <div class="editor-toolbar">
      <div class="toolbar-left">
        <span class="toolbar-label">SQL Query</span>
        <kbd>Ctrl+Enter</kbd> to run
      </div>
      <div class="toolbar-right">
        <button
          class="run-btn"
          @click="runQuery"
          :disabled="!canRun || running"
          :aria-busy="running"
        >
          <span v-if="!running">▶ Run</span>
          <span v-else>Running...</span>
        </button>
        <button
          class="outline contrast clear-btn"
          @click="clearQuery"
          title="Clear editor"
        >
          ✕
        </button>
      </div>
    </div>

    <!-- Editor area -->
    <div class="editor-wrapper" ref="editorWrapper">
      <textarea
        ref="editor"
        class="editor-textarea"
        v-model="query"
        @keydown="onKeyDown"
        @input="onInput"
        @scroll="syncScroll"
        placeholder="Type your SQL query here...
Example: SELECT * FROM table_name LIMIT 10"
        spellcheck="false"
        autocomplete="off"
        wrap="off"
      ></textarea>
      <div class="line-numbers" ref="lineNumbers">
        <div v-for="n in lineCount" :key="n" class="line-number">{{ n }}</div>
      </div>

      <!-- Autocomplete dropdown -->
      <div
        v-if="showAutocomplete"
        class="autocomplete-dropdown"
        :style="autocompleteStyle"
        ref="autocomplete"
      >
        <div
          v-for="(item, index) in filteredSuggestions"
          :key="index"
          class="autocomplete-item"
          :class="{ selected: index === autocompleteIndex }"
          @click="selectSuggestion(item)"
          @mouseenter="autocompleteIndex = index"
        >
          <span class="suggestion-type">{{ item.type }}</span>
          <span class="suggestion-value">{{ item.value }}</span>
        </div>
        <div v-if="filteredSuggestions.length === 0" class="autocomplete-empty">
          No suggestions
        </div>
      </div>
    </div>

    <!-- Linter messages -->
    <div v-if="linterMessages.length > 0" class="linter-bar">
      <div
        v-for="(msg, i) in linterMessages"
        :key="i"
        class="linter-message"
        :class="msg.type"
      >
        <span class="linter-icon">{{ msg.type === 'error' ? '✕' : '⚠' }}</span>
        {{ msg.text }}
      </div>
    </div>
  </div>
</template>

<script>
import { getAutocomplete } from '../api.js'

// SQLite keywords for autocomplete
const SQL_KEYWORDS = [
  'SELECT', 'FROM', 'WHERE', 'INSERT', 'INTO', 'VALUES', 'UPDATE', 'SET',
  'DELETE', 'CREATE', 'TABLE', 'ALTER', 'ADD', 'COLUMN', 'DROP', 'INDEX',
  'VIEW', 'TRIGGER', 'IF', 'EXISTS', 'NOT', 'NULL', 'AND', 'OR', 'IN',
  'BETWEEN', 'LIKE', 'ORDER', 'BY', 'GROUP', 'HAVING', 'LIMIT', 'OFFSET',
  'AS', 'DISTINCT', 'ALL', 'UNION', 'JOIN', 'LEFT', 'RIGHT', 'INNER',
  'OUTER', 'CROSS', 'ON', 'USING', 'CASE', 'WHEN', 'THEN', 'ELSE', 'END',
  'COUNT', 'SUM', 'AVG', 'MIN', 'MAX', 'ASC', 'DESC', 'TRUE', 'FALSE',
  'PRIMARY', 'KEY', 'FOREIGN', 'REFERENCES', 'UNIQUE', 'CHECK', 'DEFAULT',
  'AUTOINCREMENT', 'INTEGER', 'TEXT', 'REAL', 'BLOB', 'NUMERIC',
  'WITH', 'RECURSIVE', 'EXPLAIN', 'QUERY', 'PLAN', 'PRAGMA', 'CAST',
  'COALESCE', 'NULLIF', 'TYPEOF', 'LENGTH', 'SUBSTR', 'REPLACE', 'TRIM',
  'UPPER', 'LOWER', 'ABS', 'ROUND', 'RANDOM', 'DATE', 'TIME', 'DATETIME',
  'STRFTIME', 'GLOB', 'REGEXP', 'MATCH', 'ESCAPE', 'IS', 'ROWID',
]

export default {
  name: 'SqlEditor',
  props: {
    database: { type: String, default: '' },
    running: { type: Boolean, default: false },
  },
  emits: ['run-query'],
  data() {
    return {
      query: '',
      showAutocomplete: false,
      autocompleteIndex: 0,
      autocompletePosition: { top: 0, left: 0 },
      suggestions: [],
      tables: [],
      linterMessages: [],
    }
  },
  computed: {
    canRun() {
      return this.query.trim().length > 0 && !this.running
    },
    lineCount() {
      return this.query.split('\n').length
    },
    currentWord() {
      const cursorPos = this.$refs.editor?.selectionStart || 0
      const text = this.query.substring(0, cursorPos)
      const match = text.match(/(\w+)$/)
      return match ? match[1] : ''
    },
    filteredSuggestions() {
      if (!this.currentWord) return this.suggestions.slice(0, 20)
      const prefix = this.currentWord.toUpperCase()
      return this.suggestions
        .filter(s => s.value.toUpperCase().startsWith(prefix))
        .slice(0, 20)
    },
    autocompleteStyle() {
      return {
        top: `${this.autocompletePosition.top}px`,
        left: `${this.autocompletePosition.left}px`,
      }
    },
  },
  watch: {
    database: {
      immediate: true,
      handler(newDb) {
        this.loadSuggestions(newDb)
      },
    },
  },
  methods: {
    async loadSuggestions(db) {
      if (!db) {
        this.suggestions = SQL_KEYWORDS.map(k => ({ type: 'keyword', value: k }))
        return
      }

      try {
        const result = await getAutocomplete(db)
        const tableSuggestions = (result.data || []).flatMap(t => {
          const cols = (t.columns || []).map(c => ({
            type: 'column',
            value: c,
            table: t.name,
          }))
          return [{ type: 'table', value: t.name }, ...cols]
        })
        this.suggestions = [
          ...SQL_KEYWORDS.map(k => ({ type: 'keyword', value: k })),
          ...tableSuggestions,
        ]
      } catch (err) {
        console.error('Failed to load autocomplete suggestions:', err)
        this.suggestions = SQL_KEYWORDS.map(k => ({ type: 'keyword', value: k }))
      }
    },
    onKeyDown(event) {
      // Ctrl+Enter or Cmd+Enter to run
      if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
        event.preventDefault()
        this.runQuery()
        return
      }

      // Autocomplete navigation
      if (this.showAutocomplete) {
        if (event.key === 'ArrowDown') {
          event.preventDefault()
          this.autocompleteIndex = Math.min(
            this.autocompleteIndex + 1,
            this.filteredSuggestions.length - 1
          )
          return
        }
        if (event.key === 'ArrowUp') {
          event.preventDefault()
          this.autocompleteIndex = Math.max(this.autocompleteIndex - 1, 0)
          return
        }
        if (event.key === 'Enter' || event.key === 'Tab') {
          if (this.filteredSuggestions.length > 0) {
            event.preventDefault()
            this.selectSuggestion(this.filteredSuggestions[this.autocompleteIndex])
          }
          return
        }
        if (event.key === 'Escape') {
          this.showAutocomplete = false
          return
        }
      }

      // Tab for indentation
      if (event.key === 'Tab') {
        event.preventDefault()
        const start = this.$refs.editor.selectionStart
        const end = this.$refs.editor.selectionEnd
        this.query = this.query.substring(0, start) + '  ' + this.query.substring(end)
        this.$nextTick(() => {
          this.$refs.editor.selectionStart = this.$refs.editor.selectionEnd = start + 2
        })
      }
    },
    onInput() {
      this.runLinter()
      this.updateAutocomplete()
    },
    updateAutocomplete() {
      const editor = this.$refs.editor
      if (!editor) return

      const cursorPos = editor.selectionStart
      const text = this.query.substring(0, cursorPos)

      // Check if we should show autocomplete
      const wordMatch = text.match(/(\w+)$/)
      if (wordMatch && wordMatch[1].length >= 1) {
        // Calculate position for dropdown
        const line = text.substring(0, cursorPos).split('\n').length
        const lines = this.query.substring(0, cursorPos).split('\n')
        const col = lines[lines.length - 1].length

        // Approximate position based on character metrics
        const charWidth = 8.5 // approximate monospace char width
        const lineHeight = 22 // approximate line height
        this.autocompletePosition = {
          top: line * lineHeight,
          left: Math.min(col * charWidth, 400),
        }

        this.showAutocomplete = true
        this.autocompleteIndex = 0
      } else {
        this.showAutocomplete = false
      }
    },
    selectSuggestion(item) {
      const editor = this.$refs.editor
      if (!editor) return

      const cursorPos = editor.selectionStart
      const textBefore = this.query.substring(0, cursorPos)
      const textAfter = this.query.substring(cursorPos)

      // Replace the current word
      const beforeWord = textBefore.replace(/\w+$/, '')
      this.query = beforeWord + item.value + ' ' + textAfter
      this.showAutocomplete = false

      this.$nextTick(() => {
        const newPos = beforeWord.length + item.value.length + 1
        editor.selectionStart = editor.selectionEnd = newPos
        editor.focus()
      })
    },
    runQuery() {
      if (!this.canRun) return
      this.$emit('run-query', this.query.trim())
    },
    clearQuery() {
      this.query = ''
      this.linterMessages = []
      this.showAutocomplete = false
      this.$refs.editor?.focus()
    },
    runLinter() {
      const messages = []
      const q = this.query

      // Check for unmatched single quotes
      const singleQuotes = (q.match(/'/g) || []).length
      if (singleQuotes % 2 !== 0) {
        messages.push({ type: 'error', text: 'Unmatched single quote' })
      }

      // Check for unmatched double quotes
      const doubleQuotes = (q.match(/"/g) || []).length
      if (doubleQuotes % 2 !== 0) {
        messages.push({ type: 'error', text: 'Unmatched double quote' })
      }

      // Check for unmatched parentheses
      let parenCount = 0
      for (const ch of q) {
        if (ch === '(') parenCount++
        if (ch === ')') parenCount--
      }
      if (parenCount > 0) {
        messages.push({ type: 'warning', text: `${parenCount} unmatched opening parenthesis` })
      } else if (parenCount < 0) {
        messages.push({ type: 'warning', text: `${-parenCount} unmatched closing parenthesis` })
      }

      this.linterMessages = messages
    },
    syncScroll() {
      // Sync line numbers with editor scroll
      if (this.$refs.lineNumbers && this.$refs.editor) {
        this.$refs.lineNumbers.scrollTop = this.$refs.editor.scrollTop
      }
    },
  },
}
</script>

<style scoped>
.sql-editor-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.editor-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 1rem;
  border-bottom: 1px solid var(--pico-muted-border-color);
  background: var(--pico-card-background-color);
  flex-shrink: 0;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  color: var(--pico-muted-color);
}

.toolbar-label {
  font-weight: 600;
  color: var(--pico-color);
  font-size: 0.85rem;
}

.toolbar-left kbd {
  font-size: 0.7rem;
  padding: 0.1rem 0.3rem;
  background: var(--pico-card-sectioning-background-color);
  border: 1px solid var(--pico-muted-border-color);
  border-radius: 3px;
}

.toolbar-right {
  display: flex;
  gap: 0.5rem;
}

.run-btn {
  padding: 0.3rem 1rem;
  font-size: 0.85rem;
  margin-bottom: 0;
  font-weight: 600;
}

.clear-btn {
  padding: 0.3rem 0.5rem;
  font-size: 0.85rem;
  margin-bottom: 0;
}

.editor-wrapper {
  position: relative;
  flex: 1;
  overflow: hidden;
  display: flex;
}

.line-numbers {
  padding: 0.75rem 0.5rem;
  background: var(--pico-card-sectioning-background-color);
  border-right: 1px solid var(--pico-muted-border-color);
  text-align: right;
  font-family: 'SF Mono', 'Fira Code', 'Fira Mono', 'Menlo', monospace;
  font-size: 0.78rem;
  line-height: 1.6;
  color: var(--pico-muted-color);
  user-select: none;
  overflow-y: hidden;
  min-width: 3rem;
}

.line-number {
  padding: 0 0.25rem;
}

.editor-textarea {
  flex: 1;
  padding: 0.75rem 1rem;
  font-family: 'SF Mono', 'Fira Code', 'Fira Mono', 'Menlo', monospace;
  font-size: 0.85rem;
  line-height: 1.6;
  border: none;
  outline: none;
  resize: none;
  background: var(--pico-card-background-color);
  color: var(--pico-color);
  tab-size: 2;
  overflow: auto;
  min-height: 150px;
}

.editor-textarea::placeholder {
  color: var(--pico-muted-color);
  opacity: 0.6;
}

.editor-textarea:focus {
  box-shadow: none;
}

/* Autocomplete dropdown */
.autocomplete-dropdown {
  position: absolute;
  z-index: 1000;
  background: var(--pico-card-background-color);
  border: 1px solid var(--pico-muted-border-color);
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
  max-height: 240px;
  overflow-y: auto;
  min-width: 180px;
  max-width: 350px;
}

.autocomplete-item {
  padding: 0.35rem 0.75rem;
  cursor: pointer;
  display: flex;
  gap: 0.5rem;
  align-items: center;
  font-size: 0.8rem;
  font-family: 'SF Mono', 'Fira Code', 'Fira Mono', monospace;
}

.autocomplete-item:hover,
.autocomplete-item.selected {
  background: var(--pico-primary);
  color: var(--pico-primary-inverse);
}

.suggestion-type {
  font-size: 0.65rem;
  padding: 0.05rem 0.3rem;
  border-radius: 3px;
  background: var(--pico-muted-border-color);
  text-transform: uppercase;
  font-weight: 600;
  flex-shrink: 0;
}

.autocomplete-item.selected .suggestion-type {
  background: rgba(255,255,255,0.2);
}

.suggestion-value {
  overflow: hidden;
  text-overflow: ellipsis;
}

.autocomplete-empty {
  padding: 0.5rem 0.75rem;
  color: var(--pico-muted-color);
  font-size: 0.8rem;
  text-align: center;
}

/* Linter bar */
.linter-bar {
  border-top: 1px solid var(--pico-muted-border-color);
  background: var(--pico-card-sectioning-background-color);
  padding: 0.25rem 1rem;
  flex-shrink: 0;
  max-height: 60px;
  overflow-y: auto;
}

.linter-message {
  font-size: 0.78rem;
  padding: 0.15rem 0;
  display: flex;
  gap: 0.4rem;
  align-items: center;
}

.linter-message.error {
  color: var(--pico-del-color);
}

.linter-message.warning {
  color: var(--pico-warning);
}

.linter-icon {
  font-weight: bold;
}
</style>
