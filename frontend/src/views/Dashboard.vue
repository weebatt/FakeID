<template>
  <div class="dashboard">
    <div class="stats">
      <div class="stat-card">
        <h3>Total Requests</h3>
        <p class="stat-value">12</p>
      </div>
      <div class="stat-card">
        <h3>Generated Tests</h3>
        <p class="stat-value">34</p>
      </div>
      <div class="stat-card">
        <h3>Successful Tests</h3>
        <p class="stat-value">28</p>
      </div>
    </div>

    <div class="quick-actions">
      <h2>Quick Actions</h2>
      <div class="action-buttons">
        <button class="accent-button" @click="$router.push('/request-builder')">Create New Request</button>
        <button @click="$router.push('/test-results')">View Test Results</button>
      </div>
    </div>

    <div class="recent-activity">
      <h2>Recent Activity</h2>
      <div class="activity-list">
        <div class="activity-item" v-for="(activity, index) in recentActivity" :key="index">
          <div class="activity-icon" :class="activity.type">{{ activity.icon }}</div>
          <div class="activity-details">
            <h4>{{ activity.title }}</h4>
            <p>{{ activity.description }}</p>
            <span class="activity-time">{{ activity.time }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Dashboard',
  data() {
    return {
      recentActivity: [
        {
          type: 'request',
          icon: '🔄',
          title: 'GET Users API',
          description: 'Test data generated successfully',
          time: '10 minutes ago'
        },
        {
          type: 'test',
          icon: '✅',
          title: 'POST Product API',
          description: 'All tests passed',
          time: '2 hours ago'
        },
        {
          type: 'request',
          icon: '🔄',
          title: 'PUT User API',
          description: 'Test data generated with warnings',
          time: '5 hours ago'
        }
      ]
    }
  }
}
</script>

<style scoped>
.dashboard {
  padding: 24px;
}

.stats {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 32px;
}

.stat-card {
  background-color: var(--secondary-color);
  padding: 16px;
  border-radius: 4px;
  text-align: center;
}

.stat-value {
  font-size: 36px;
  font-weight: bold;
  color: var(--accent-color);
}

h2 {
  margin-bottom: 16px;
  font-size: 18px;
}

.quick-actions {
  margin-bottom: 32px;
}

.action-buttons {
  display: flex;
  gap: 16px;
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.activity-item {
  display: flex;
  background-color: var(--secondary-color);
  padding: 16px;
  border-radius: 4px;
}

.activity-icon {
  font-size: 24px;
  margin-right: 16px;
  padding: 8px;
  border-radius: 50%;
  background-color: rgba(0, 0, 0, 0.2);
  height: 24px;
  width: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.activity-time {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
}
</style>