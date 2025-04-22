<template>
  <div class="auth-container">
    <div class="auth-card">
      <div class="auth-header">
        <h2>API Test Generator</h2>
        <p>Create a new account</p>
      </div>

      <div class="auth-form">
        <div class="form-group" :class="{ 'error': errors.name }">
          <label for="name">Full Name</label>
          <input
              type="text"
              id="name"
              v-model="form.name"
              placeholder="Enter your full name"
              @input="clearError('name')"
          />
          <p class="error-message" v-if="errors.name">{{ errors.name }}</p>
        </div>

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
              placeholder="Create a password"
              @input="clearError('password')"
          />
          <p class="error-message" v-if="errors.password">{{ errors.password }}</p>
        </div>

        <div class="form-group" :class="{ 'error': errors.confirmPassword }">
          <label for="confirmPassword">Confirm Password</label>
          <input
              type="password"
              id="confirmPassword"
              v-model="form.confirmPassword"
              placeholder="Confirm your password"
              @input="clearError('confirmPassword')"
          />
          <p class="error-message" v-if="errors.confirmPassword">{{ errors.confirmPassword }}</p>
        </div>

        <div class="terms-agreement">
          <input type="checkbox" id="termsAgree" v-model="form.termsAgree" />
          <label for="termsAgree">
            I agree to the <a href="#" @click.prevent="showTerms">Terms of Service</a> and
            <a href="#" @click.prevent="showPrivacy">Privacy Policy</a>
          </label>
          <p class="error-message" v-if="errors.termsAgree">{{ errors.termsAgree }}</p>
        </div>

        <button
            class="auth-button accent-button"
            @click="register"
            :disabled="isLoading"
        >
          {{ isLoading ? 'Creating Account...' : 'Create Account' }}
        </button>

        <p class="auth-error" v-if="authError">{{ authError }}</p>
      </div>

      <div class="auth-footer">
        <p>Already have an account? <router-link to="/login">Log in</router-link></p>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Register',
  data() {
    return {
      form: {
        name: '',
        email: '',
        password: '',
        confirmPassword: '',
        termsAgree: false
      },
      errors: {
        name: '',
        email: '',
        password: '',
        confirmPassword: '',
        termsAgree: ''
      },
      authError: '',
      isLoading: false
    }
  },
  methods: {
    register() {
      // Reset errors
      this.errors = {
        name: '',
        email: '',
        password: '',
        confirmPassword: '',
        termsAgree: ''
      };
      this.authError = '';

      // Validate form
      let isValid = true;

      if (!this.form.name) {
        this.errors.name = 'Name is required';
        isValid = false;
      }

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
      } else if (this.form.password.length < 8) {
        this.errors.password = 'Password must be at least 8 characters';
        isValid = false;
      }

      if (!this.form.confirmPassword) {
        this.errors.confirmPassword = 'Please confirm your password';
        isValid = false;
      } else if (this.form.password !== this.form.confirmPassword) {
        this.errors.confirmPassword = 'Passwords do not match';
        isValid = false;
      }

      if (!this.form.termsAgree) {
        this.errors.termsAgree = 'You must agree to the terms and privacy policy';
        isValid = false;
      }

      if (!isValid) return;

      // Set loading state
      this.isLoading = true;

      // Simulate API call
      setTimeout(() => {
        // For demo purposes, check if email already exists
        if (this.form.email === 'demo@example.com') {
          this.authError = 'This email is already registered';
        } else {
          // Store user info (in a real app, this would be done after API confirmation)
          localStorage.setItem('user', JSON.stringify({
            id: 2,
            email: this.form.email,
            name: this.form.name,
            token: 'fake-jwt-token'
          }));

          // Redirect to dashboard
          this.$router.push('/');
        }

        this.isLoading = false;
      }, 1500);
    },
    validateEmail(email) {
      const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      return re.test(email);
    },
    clearError(field) {
      this.errors[field] = '';
      this.authError = '';
    },
    showTerms() {
      alert('Terms of Service would be shown here');
    },
    showPrivacy() {
      alert('Privacy Policy would be shown here');
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

.terms-agreement {
  display: flex;
  align-items: flex-start;
  margin-bottom: 24px;
  font-size: 14px;
}

.terms-agreement input {
  width: auto;
  margin-right: 8px;
  margin-top: 3px;
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
