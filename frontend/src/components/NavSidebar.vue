<template>
  <div class="sidebar">
    <div class="logo">
      <h2>API Test Generator</h2>
    </div>
    <nav>
      <div class="nav-section">
        <h3>Project</h3>
        <ul>
          <li><router-link to="/">Dashboard</router-link></li>
          <li><router-link to="/request-builder">Request Builder</router-link></li>
          <li><router-link to="/test-results">Test Results</router-link></li>
        </ul>
      </div>
      <div class="nav-section">
        <h3>Recent Requests</h3>
        <ul>
          <li v-for="(request, index) in recentRequests" :key="index">
            <a href="#" @click.prevent="loadRequest(request)">{{ request.name }}</a>
          </li>
        </ul>
      </div>
    </nav>
  </div>
</template>

<script>
export default {
  name: 'NavSidebar',
  data() {
    return {
      recentRequests: [
        { name: 'GET Users', method: 'GET', url: 'https://api.example.com/users' },
        { name: 'POST User', method: 'POST', url: 'https://api.example.com/users' },
        { name: 'GET Products', method: 'GET', url: 'https://api.example.com/products' }
      ]
    }
  },
  methods: {
    loadRequest(request) {
      this.$router.push('/request-builder');
      this.$emit('load-request', request);
    }
  }
}
</script>

<style scoped>
.sidebar {
  width: 250px;
  background-color: var(--sidebar-color);
  color: var(--text-color);
  border-right: 1px solid var(--border-color);
  height: 100vh;
  overflow-y: auto;
  position: sticky;
  top: 0;
  padding: 16px 0;
}

.logo {
  padding: 0 16px 16px;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 16px;
}

.nav-section {
  padding: 0 16px;
  margin-bottom: 24px;
}

.nav-section h3 {
  margin-bottom: 8px;
  font-size: 16px;
  color: var(--accent-color);
}

ul {
  list-style: none;
}

li {
  margin-bottom: 8px;
}

a {
  display: block;
  padding: 8px 8px;
  border-radius: 4px;
}

a:hover, a.router-link-active {
  background-color: rgba(255, 255, 255, 0.1);
}
</style>