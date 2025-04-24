<template>
  <header class="app-header">
    <div class="title">
      <h1>{{ currentPage }}</h1>
    </div>
    <div class="actions">
      <button class="accent-button" @click="exportData">Export Data</button>
      <button>Settings</button>
      <button @click="signOut">Sign Out</button>
    </div>
  </header>
</template>

<script>
export default {
  name: 'AppHeader',
  computed: {
    currentPage() {
      const route = this.$route.path;
      if (route === '/') return 'Dashboard';
      if (route === '/request-builder') return 'Request Builder';
      if (route === '/test-results') return 'Test Results';
      return 'API Test Generator';
    }
  },
  methods: {
    exportData() {
      // Logic to export data as ZIP
      alert('Exporting data as ZIP...');
    },
    signOut() {
      // Удаляем пользовательские данные из localStorage
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      // При необходимости очистить ещё что-то:
      // localStorage.clear(); // полный сброс всех данных

      // Перенаправляем пользователя на страницу входа
      this.$router.push({ name: 'Login' });
    }
  }
}
</script>

<style scoped>
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color);
}

.title h1 {
  font-size: 20px;
}

.actions {
  display: flex;
  gap: 8px;
}
</style>
