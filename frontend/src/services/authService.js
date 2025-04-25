// This would normally connect to your backend API

const authService = {
    // Simulate login API call
    async login(email, password) {
        return new Promise((resolve, reject) => {
            setTimeout(() => {
                // For demo purposes
                if (email === 'demo@example.com' && password === 'password') {
                    resolve({
                        id: 1,
                        name: 'Demo User',
                        email: 'demo@example.com',
                        token: 'fake-jwt-token'
                    });
                } else {
                    reject(new Error('Invalid email or password'));
                }
            }, 1000);
        });
    },

    // Simulate register API call
    async register(name, email, password) {
        return new Promise((resolve, reject) => {
            setTimeout(() => {
                // For demo purposes, check if email already exists
                if (email === 'demo@example.com') {
                    reject(new Error('This email is already registered'));
                } else {
                    resolve({
                        id: Math.floor(Math.random() * 1000) + 2, // Random ID
                        name,
                        email,
                        token: 'fake-jwt-token'
                    });
                }
            }, 1500);
        });
    },

    // Simulate forgot password API call
    async forgotPassword(email) {
        return new Promise((resolve) => {
            setTimeout(() => {
                // Always succeed for demo purposes
                resolve(true);
            }, 1500);
        });
    },

    // Verify token validity
    async verifyToken(token) {
        return new Promise((resolve) => {
            setTimeout(() => {
                // For demo purposes, any token is valid
                resolve(true);
            }, 500);
        });
    }
};

export default authService;