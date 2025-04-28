<template>
  <header class="app-header">
    <div class="title">
      <h1>{{ currentPage }}</h1>
    </div>
    <div class="actions">
      <button class="accent-button" @click="exportData">Export Data</button>
      <button>Settings</button>
      <button @click="logOut">Log Out</button>
    </div>
  </header>
</template>

<script>
import router from "@/router/index.js";

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
      alert('Exporting data as ZIP...');
    },
    logOut() {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      this.$router.push('/login').then(() => {
        window.location.reload();
      });
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
