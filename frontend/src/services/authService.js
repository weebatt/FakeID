// services/authService.js - Обновлен для работы через API Gateway

const API_URL = '/api/auth'; // Относительный путь, который будет проксироваться через Nginx

const authService = {
    // Авторизация через API Gateway
    async login(email, password) {
        try {
            const response = await fetch(`${API_URL}/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Ошибка входа');
            }

            const data = await response.json();
            return data.data; // API Gateway оборачивает ответ в поле data
        } catch (error) {
            throw error;
        }
    },

    // Регистрация через API Gateway
    async register(name, email, password) {
        try {
            const response = await fetch(`${API_URL}/register`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ name, email, password })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Ошибка регистрации');
            }

            const data = await response.json();
            return data.data;
        } catch (error) {
            throw error;
        }
    },

    // Восстановление пароля через API Gateway
    async forgotPassword(email) {
        try {
            const response = await fetch(`${API_URL}/forgot-password`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Не удалось отправить ссылку для сброса пароля');
            }

            const data = await response.json();
            return data.data;
        } catch (error) {
            throw error;
        }
    },

    // Проверка токена с использованием заголовка Authentication
    async verifyToken(token) {
        try {
            const response = await fetch(`${API_URL}/verify-token`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                }
            });

            return response.ok;
        } catch (error) {
            return false;
        }
    }
};

export default authService;