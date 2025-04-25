// store/auth.js
import { reactive } from 'vue';
import router from '../router';
import authService from '../services/authService';

// Create a reactive store
const state = reactive({
    user: null,
    token: null,
    isAuthenticated: false,
    isLoading: false,
    error: null
});

// Initialize state from localStorage
const initState = () => {
    const storedUser = localStorage.getItem('user');
    const storedToken = localStorage.getItem('token');
    if (storedUser && storedToken) {
        try {
            state.user = JSON.parse(storedUser);
            state.token = storedToken;
            state.isAuthenticated = true;
        } catch (e) {
            localStorage.removeItem('user');
            localStorage.removeItem('token');
        }
    }
};

// Auth actions
const actions = {
    async login(email, password, remember = false) {
        state.isLoading = true;
        state.error = null;

        try {
            const response = await authService.login(email, password);
            console.log('Login response:', response);

            // Проверяем, что response содержит user и token
            if (!response || !response.user || !response.token) {
                throw new Error('Invalid response from server: missing user or token');
            }

            state.user = response.user;
            state.token = response.token;
            state.isAuthenticated = true;

            localStorage.setItem('user', JSON.stringify(response.user));
            localStorage.setItem('token', response.token);

            if (remember) {
                localStorage.setItem('rememberMe', 'true');
            }

            router.push('/');
            return response.user;
        } catch (error) {
            console.error('Login error:', error);
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
            const response = await authService.register(name, email, password);
            console.log('Register response:', response);

            // Проверяем, что response содержит user и token
            if (!response || !response.user || !response.token) {
                throw new Error('Invalid response from server: missing user or token');
            }

            state.user = response.user;
            state.token = response.token;
            state.isAuthenticated = true;

            localStorage.setItem('user', JSON.stringify(response.user));
            localStorage.setItem('token', response.token);

            router.push('/');
            return response.user;
        } catch (error) {
            console.error('Register error:', error);
            state.error = error.message || 'Registration failed';
            throw error;
        } finally {
            state.isLoading = false;
        }
    },

    logout() {
        state.user = null;
        state.token = null;
        state.isAuthenticated = false;
        localStorage.removeItem('user');
        localStorage.removeItem('token');
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