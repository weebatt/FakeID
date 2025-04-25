const API_URL = '/api/v1';

const authService = {
    async login(email, password) {
        try {
            console.log('Sending login request to:', `${API_URL}/login`, 'with body:', { email, password });
            const response = await fetch(`${API_URL}/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password })
            });

            console.log('Login response status:', response.status, 'OK:', response.ok);

            if (!response.ok) {
                const errorData = await response.json();
                console.log('Login error response:', errorData);
                throw new Error(errorData.error || 'Ошибка входа');
            }

            const data = await response.json();
            console.log('Login response data:', data);

            // Преобразуем данные в ожидаемый формат
            const result = {
                user: data.data?.user || { email: data.data?.email, name: data.data?.name || '' },
                token: data.data?.token,
            };

            if (!result.user || !result.token) {
                throw new Error('Response missing user or token');
            }

            return result;
        } catch (error) {
            console.error('Login fetch error:', error);
            throw error;
        }
    },

    async register(name, email, password) {
        try {
            console.log('Sending register request to:', `${API_URL}/register`, 'with body:', { name, email, password });
            const response = await fetch(`${API_URL}/register`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password })
            });

            console.log('Register response status:', response.status, 'OK:', response.ok);

            if (!response.ok) {
                const errorData = await response.json();
                console.log('Register error response:', errorData);
                throw new Error(errorData.error || 'Ошибка регистрации');
            }

            const data = await response.json();
            console.log('Register response data:', data);

            // Преобразуем данные в ожидаемый формат
            const result = {
                user: data.data?.user || { email: data.data?.email, name: data.data?.name || name || '' },
                token: data.data?.token,
            };

            if (!result.user || !result.token) {
                throw new Error('Response missing user or token');
            }

            return result;
        } catch (error) {
            console.error('Register fetch error:', error);
            throw error;
        }
    },
};

export default authService;