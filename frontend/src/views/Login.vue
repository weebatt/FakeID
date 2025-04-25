<template>
  <div class="auth-container">
    <div class="auth-card">
      <div class="auth-header">
        <h2>API Test Generator</h2>
        <p>Log in to your account</p>
      </div>

      <div class="auth-form">
        <div class="form-group" :class="{ 'error': errors.email }">
          <label for="email">Email</label>
          <input
              type="email"
              id="email"
              v-model="form.email"
              placeholder="Enter your email"
              @input="clearError('email')"
          />
          <p class="error-message" v-if="errors.email">{{ errors.email }}</p>
        </div>

        <div class="form-group" :class="{ 'error': errors.password }">
          <label for="password">Password</label>
          <input
              type="password"
              id="password"
              v-model="form.password"
              placeholder="Enter your password"
              @input="clearError('password')"
          />
          <p class="error-message" v-if="errors.password">{{ errors.password }}</p>
        </div>

        <div class="form-actions">
          <div class="remember-me">
            <input type="checkbox" id="remember" v-model="form.remember" />
            <label for="remember">Remember me</label>
          </div>
          <a href="#" @click.prevent="forgotPassword">Forgot password?</a>
        </div>

        <button
            class="auth-button accent-button"
            @click="login"
            :disabled="isLoading"
        >
          {{ isLoading ? 'Logging in...' : 'Log In' }}
        </button>

        <p class="auth-error" v-if="authError">{{ authError }}</p>
      </div>

      <div class="auth-footer">
        <p>Don't have an account? <router-link to="/register">Sign up</router-link></p>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Login',
  data() {
    return {
      form: {
        email: '',
        password: '',
        remember: false
      },
      errors: {
        email: '',
        password: ''
      },
      authError: '',
      isLoading: false
    }
  },
  methods: {
    login() {
      // Reset errors
      this.errors = {
        email: '',
        password: ''
      };
      this.authError = '';

      // Validate form
      let isValid = true;

      if (!this.form.email) {
        this.errors.email = 'Email is required';
        isValid = false;
      } else if (!this.validateEmail(this.form.email)) {
        this.errors.email = 'Please enter a valid email';
        isValid = false;
      }

      if (!this.form.password) {
        this.errors.password = 'Password is required';
        isValid = false;
      }

      if (!isValid) return;

      // Set loading state
      this.isLoading = true;

      // Simulate API call
      setTimeout(() => {
        // For demo purposes, let's simulate a successful login
        if (this.form.email === 'demo@example.com' && this.form.password === 'password') {
          // Store user info in localStorage or Vuex store
          localStorage.setItem('user', JSON.stringify({
            id: 1,
            email: this.form.email,
            name: 'Demo User',
            token: 'fake-jwt-token'
          }));

          // Redirect to dashboard
          this.$router.push('/');
        } else {
          // Show error
          this.authError = 'Invalid email or password';
        }

        this.isLoading = false;
      }, 1000);
    },
    validateEmail(email) {
      const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      return re.test(email);
    },
    clearError(field) {
      this.errors[field] = '';
      this.authError = '';
    },
    forgotPassword() {
      alert('Password reset functionality would be implemented here');
    }
  }
}
</script>

<style scoped>
.auth-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 24px;
  background-color: var(--background-color);
}

.auth-card {
  background-color: var(--secondary-color);
  border-radius: 8px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
}

.auth-header {
  padding: 24px;
  text-align: center;
  border-bottom: 1px solid var(--border-color);
}

.auth-header h2 {
  margin-bottom: 8px;
}

.auth-header p {
  color: rgba(255, 255, 255, 0.7);
}

.auth-form {
  padding: 24px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
}

.form-group.error input {
  border-color: #f44336;
}

.error-message {
  color: #f44336;
  font-size: 12px;
  margin-top: 4px;
}

.form-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  font-size: 14px;
}

.remember-me {
  display: flex;
  align-items: center;
}

.remember-me input {
  width: auto;
  margin-right: 8px;
}

.auth-button {
  width: 100%;
  padding: 12px;
  font-size: 16px;
}

.auth-error {
  color: #f44336;
  text-align: center;
  margin-top: 16px;
}

.auth-footer {
  padding: 16px 24px;
  text-align: center;
  border-top: 1px solid var(--border-color);
  font-size: 14px;
}
</style>