// store/auth.js
import { reactive } from 'vue';
import router from '../router';
import authService from '../services/authService';

// Create a reactive store
const state = reactive({
    user: null,
    isAuthenticated: false,
    isLoading: false,
    error: null
});

// Initialize state from localStorage
const initState = () => {
    const storedUser = localStorage.getItem('user');
    if (storedUser) {
        try {
            state.user = JSON.parse(storedUser);
            state.isAuthenticated = true;
        } catch (e) {
            localStorage.removeItem('user');
        }
    }
};

// Auth actions
const actions = {
    async login(email, password, remember = false) {
        state.isLoading = true;
        state.error = null;

        try {
            const user = await authService.login(email, password);
            state.user = user;
            state.isAuthenticated = true;
            localStorage.setItem('user', JSON.stringify(user));

            if (remember) {
                // Additional logic for remember me
                localStorage.setItem('rememberMe', 'true');
            }

            router.push('/');
            return user;
        } catch (error) {
            state.error = error.message || 'Login failed';
            throw error;
        } finally {
            state.isLoading = false;
        }
    },

    async register(name, email, password) {
        state.isLoading = true;
        state.error = null;

        try {
            const user = await authService.register(name, email, password);
            state.user = user;
            state.isAuthenticated = true;
            localStorage.setItem('user', JSON.stringify(user));

            router.push('/');
            return user;
        } catch (error) {
            state.error = error.message || 'Registration failed';
            throw error;
        } finally {
            state.isLoading = false;
        }
    },

    async forgotPassword(email) {
        state.isLoading = true;
        state.error = null;

        try {
            await authService.forgotPassword(email);
            return true;
        } catch (error) {
            state.error = error.message || 'Failed to send reset link';
            throw error;
        } finally {
            state.isLoading = false;
        }
    },

    logout() {
        state.user = null;
        state.isAuthenticated = false;
        localStorage.removeItem('user');
        localStorage.removeItem('rememberMe');
        router.push('/login');
    },

    clearError() {
        state.error = null;
    }
};

// Initialize on creation
initState();

// Export the auth store
export default {
    state,
    ...actions
};
