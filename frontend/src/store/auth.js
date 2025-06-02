import { reactive } from 'vue';
import router from '../router';
import authService from '../services/authService';

const state = reactive({
    user: null,
    token: null,
    isAuthenticated: false,
    isLoading: false,
    error: null
});

const initState = () => {
    const storedUser = localStorage.getItem('user');
    const storedToken = localStorage.getItem('token');
    if (storedUser && storedToken) {
        try {
            state.user = JSON.parse(storedUser);
            state.token = storedToken;
            state.isAuthenticated = true;
        } catch (e) {
            // localStorage.removeItem('user');
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

            if (!response || !response.token) {
                throw new Error('Invalid response from server: missing user or token');
            }

            state.user = response.user;
            state.token = response.token;
            state.isAuthenticated = true;

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

            if (!response.ok) {
                throw new Error('Invalid response from server');
            }

            state.message = response.message;
            state.user_id = response.user_id;

            router.push('/');
            return response.message;
        } catch (error) {
            console.error('Register error:', error);
            state.error = error.message || 'Registration failed';
            throw error;
        } finally {
            state.isLoading = false;
        }
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