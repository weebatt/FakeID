<template>
  <div class="auth-container">
    <div class="auth-card">
      <div class="auth-header">
        <h2>Reset Password</h2>
        <p>Enter your email to receive a reset link</p>
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

        <button
            class="auth-button accent-button"
            @click="resetPassword"
            :disabled="isLoading || isSuccess"
        >
          {{ isLoading ? 'Sending...' : (isSuccess ? 'Email Sent' : 'Send Reset Link') }}
        </button>

        <p class="auth-error" v-if="authError">{{ authError }}</p>
        <p class="auth-success" v-if="isSuccess">
          Password reset instructions have been sent to your email
        </p>
      </div>

      <div class="auth-footer">
        <p><router-link to="/login">Back to Login</router-link></p>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ForgotPassword',
  data() {
    return {
      form: {
        email: ''
      },
      errors: {
        email: ''
      },
      authError: '',
      isLoading: false,
      isSuccess: false
    }
  },
  methods: {
    resetPassword() {
      // Reset errors
      this.errors = {
        email: ''
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

      if (!isValid) return;

      // Set loading state
      this.isLoading = true;

      // Simulate API call
      setTimeout(() => {
        // For demo purposes, always success
        this.isSuccess = true;
        this.isLoading = false;

        // After some time, redirect to login
        setTimeout(() => {
          this.$router.push('/login');
        }, 3000);
      }, 1500);
    },
    validateEmail(email) {
      const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      return re.test(email);
    },
    clearError(field) {
      this.errors[field] = '';
      this.authError = '';
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

.auth-success {
  color: #4caf50;
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