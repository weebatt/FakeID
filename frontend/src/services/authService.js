const API_URL = '/api/v1';

const authService = {
    async login(email, password) {
        const response = await fetch(`${API_URL}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        if (!response.ok) {
            const err = await response.json();
            throw new Error(err.error || 'Ошибка входа');
        }

        const data = await response.json();
        console.log('Login response data:', data);

        // Берём данные либо из data.user/token, либо из data.data.user/token
        const user = data.user ?? data.data?.user;
        const token = data.token ?? data.data?.token;

        console.log("token", token);
        console.log("user", user);

        if (!token) {
            throw new Error('Response missing user or token');
        }

        return { user, token };
    },

    async register(name, email, password) {
        const response = await fetch(`${API_URL}/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            // ОБЯЗАТЕЛЬНО отправляем name, иначе сервер может не вернуть его
            body: JSON.stringify({ name, email, password })
        });

        if (!response.ok) {
            const err = await response.json();
            throw new Error(err.error || 'Ошибка регистрации');
        }

        const data = await response.json();
        console.log('Register response data:', data);

        const user = data.user ?? data.data?.user;
        const token = data.token ?? data.data?.token;

        if (!token) {
            throw new Error('Response missing user or token');
        }

        return { user, token };
    }
};

export default authService;