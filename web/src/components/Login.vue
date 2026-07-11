<template>
  <div class="login-overlay">
    <div class="login-card">
      <div class="login-header">
        <span class="login-icon">📊</span>
        <h1>DBQuery</h1>
        <p class="login-subtitle">{{ isRegister ? 'Create your account' : 'Sign in to your account' }}</p>
      </div>

      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label for="username">Username</label>
          <input
            id="username"
            v-model="username"
            type="text"
            placeholder="Enter username"
            autocomplete="username"
            required
            minlength="3"
          />
        </div>

        <div class="form-group">
          <label for="password">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            placeholder="Enter password"
            autocomplete="current-password"
            required
            minlength="6"
          />
        </div>

        <div v-if="error" class="login-error">
          {{ error }}
        </div>

        <button
          type="submit"
          class="login-btn"
          :disabled="loading || !username || !password"
          :aria-busy="loading"
        >
          <span v-if="!loading">{{ isRegister ? 'Create Account' : 'Sign In' }}</span>
          <span v-else>{{ isRegister ? 'Creating...' : 'Signing in...' }}</span>
        </button>
      </form>

      <div class="login-footer" v-if="!isRegister && !hasUsers">
        <p>No account found. <a href="#" @click.prevent="isRegister = true">Create the first account</a></p>
      </div>
    </div>
  </div>
</template>

<script>
import { hasUsers, login, register } from '../api.js'

export default {
  name: 'LoginView',
  emits: ['auth-success'],
  data() {
    return {
      username: '',
      password: '',
      isRegister: false,
      hasUsers: true,
      loading: false,
      error: '',
    }
  },
  async mounted() {
    // Check if any users exist to determine login vs register mode
    try {
      const result = await hasUsers()
      this.hasUsers = result.data?.has_users ?? true
      if (!this.hasUsers) {
        this.isRegister = true
      }
    } catch (err) {
      // If we can't check, assume login mode
      console.error('Failed to check users:', err)
    }
  },
  methods: {
    async handleSubmit() {
      this.error = ''
      this.loading = true

      try {
        let result
        if (this.isRegister) {
          result = await register(this.username, this.password)
        } else {
          result = await login(this.username, this.password)
        }

        // Store token and username
        const { token, username } = result.data
        localStorage.setItem('auth_token', token)
        localStorage.setItem('auth_username', username)

        this.$emit('auth-success', { token, username })
      } catch (err) {
        this.error = err.message || 'Authentication failed'
      } finally {
        this.loading = false
      }
    },
  },
}
</script>

<style scoped>
.login-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--pico-background-color);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.login-card {
  width: 100%;
  max-width: 400px;
  padding: 2rem;
  background: var(--pico-card-background-color);
  border: 1px solid var(--pico-muted-border-color);
  border-radius: 12px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.1);
}

.login-header {
  text-align: center;
  margin-bottom: 1.5rem;
}

.login-icon {
  font-size: 2.5rem;
  display: block;
  margin-bottom: 0.5rem;
}

.login-header h1 {
  font-size: 1.75rem;
  font-weight: 700;
  margin: 0;
  color: var(--pico-primary);
}

.login-subtitle {
  color: var(--pico-muted-color);
  margin: 0.5rem 0 0;
  font-size: 0.95rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.35rem;
  font-weight: 600;
  font-size: 0.9rem;
}

.form-group input {
  width: 100%;
  padding: 0.6rem 0.75rem;
  font-size: 1rem;
  margin-bottom: 0;
}

.login-error {
  background: #fef2f2;
  color: #dc2626;
  padding: 0.6rem 0.75rem;
  border-radius: 6px;
  font-size: 0.875rem;
  margin-bottom: 1rem;
  border: 1px solid #fecaca;
}

.login-btn {
  width: 100%;
  padding: 0.65rem;
  font-size: 1rem;
  font-weight: 600;
  margin-top: 0.5rem;
  margin-bottom: 0;
}

.login-footer {
  text-align: center;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--pico-muted-border-color);
}

.login-footer p {
  margin: 0;
  font-size: 0.875rem;
  color: var(--pico-muted-color);
}

.login-footer a {
  color: var(--pico-primary);
  text-decoration: none;
  font-weight: 600;
}

.login-footer a:hover {
  text-decoration: underline;
}
</style>
