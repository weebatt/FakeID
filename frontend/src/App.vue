<template>
  <div class="app-container">
    <NavSidebar />
    <div class="main-content">
      <AppHeader />
      <router-view />
    </div>
  </div>
</template>

<script>
import NavSidebar from './components/NavSidebar.vue';
import AppHeader from './components/AppHeader.vue';

export default {
  name: 'App',
  components: {
    NavSidebar,
    AppHeader
  }
}
</script>

<style>
:root {
  --background-color: #1a1a1a;
  --sidebar-color: #131313;
  --text-color: #e0e0e0;
  --accent-color: #4caf50;
  --secondary-color: #424242;
  --border-color: #333333;
  --input-bg-color: #222222;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
  font-family: 'Courier New', monospace;
}

body {
  background-color: var(--background-color);
  color: var(--text-color);
}

.app-container {
  display: flex;
  min-height: 100vh;
}

.main-content {
  flex: 1;
  padding: 0;
  overflow-y: auto;
}

a {
  color: var(--accent-color);
  text-decoration: none;
}

button {
  cursor: pointer;
  background-color: var(--secondary-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
  padding: 8px 16px;
  border-radius: 4px;
  font-size: 14px;
}

button:hover {
  background-color: var(--border-color);
}

.accent-button {
  background-color: var(--accent-color);
  color: #000;
}

.accent-button:hover {
  background-color: #3d8b40;
}

input, textarea, select {
  background-color: var(--input-bg-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
  padding: 8px;
  border-radius: 4px;
  font-size: 14px;
  width: 100%;
}

.card {
  background-color: var(--secondary-color);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 16px;
  margin-bottom: 16px;
}
</style>