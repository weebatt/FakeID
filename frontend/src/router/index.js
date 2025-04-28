// router/index.js (updated)
import { createRouter, createWebHistory } from 'vue-router';
import Dashboard from '../views/Dashboard.vue';
import RequestBuilder from '../views/RequestBuilder.vue';
import TestResults from '../views/TestResults.vue';
import Login from '../views/Login.vue';
import Register from '../views/Register.vue';
import ForgotPassword from '../views/ForgotPassword.vue';
import UserProfile from '../views/UserProfile.vue';
import authStore from '../store/auth';

const routes = [
    {
        path: '/',
        name: 'Dashboard',
        component: Dashboard,
        meta: { requiresAuth: true }
    },
    {
        path: '/request-builder',
        name: 'RequestBuilder',
        component: RequestBuilder,
        meta: { requiresAuth: true }
    },
    {
        path: '/test-results',
        name: 'TestResults',
        component: TestResults,
        meta: { requiresAuth: true }
    },
    {
        path: '/profile',
        name: 'UserProfile',
        component: UserProfile,
        meta: { requiresAuth: true }
    },
    {
        path: '/login',
        name: 'Login',
        component: Login,
        meta: { guest: true }
    },
    {
        path: '/register',
        name: 'Register',
        component: Register,
        meta: { guest: true }
    },
    {
        path: '/forgot-password',
        name: 'ForgotPassword',
        component: ForgotPassword,
        meta: { guest: true }
    }
];

const router = createRouter({
    history: createWebHistory(),
    routes
});

// Navigation guard
router.beforeEach((to, from, next) => {
    const isAuthenticated = authStore.state.isAuthenticated;

    // Routes that require authentication
    if (to.matched.some(record => record.meta.requiresAuth)) {
        if (!isAuthenticated) {
            next({ name: 'Login', query: { redirect: to.fullPath } });
        } else {
            next();
        }
    }
    // Routes for guests only (prevent logged in users from accessing login/register)
    else if (to.matched.some(record => record.meta.guest)) {
        if (isAuthenticated) {
            next({ name: 'Dashboard' });
        } else {
            next();
        }
    }
    // Public routes
    else {
        next();
    }
});

export default router;