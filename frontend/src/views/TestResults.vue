<template>
  <div class="test-results">
    <div class="test-summary">
      <div class="summary-card">
        <h3>Total Tests</h3>
        <p class="stat-value">{{ testSummary.total }}</p>
      </div>
      <div class="summary-card">
        <h3>Passed</h3>
        <p class="stat-value success">{{ testSummary.passed }}</p>
      </div>
      <div class="summary-card">
        <h3>Failed</h3>
        <p class="stat-value error">{{ testSummary.failed }}</p>
      </div>
      <div class="summary-card">
        <h3>Success Rate</h3>
        <p class="stat-value">{{ testSummary.successRate }}%</p>
      </div>
    </div>

    <div class="results-filters">
      <div class="filter-item">
        <label>API Endpoint</label>
        <select v-model="filters.endpoint">
          <option value="">All Endpoints</option>
          <option>POST /api/v2/tasks</option>
          <option>GET /api/v2/tasks/:id</option>
        </select>
      </div>
      <div class="filter-item">
        <label>Status</label>
        <select v-model="filters.status">
          <option value="">All</option>
          <option value="passed">Passed</option>
          <option value="failed">Failed</option>
        </select>
      </div>
      <div class="filter-item">
        <label>Date</label>
        <input type="date" v-model="filters.date" />
      </div>
      <button class="filter-apply" @click="applyFilters">Apply Filters</button>
    </div>

    <div class="test-list">
      <div class="test-item" v-for="(test, index) in filteredTests" :key="test.id">
        <div class="test-header" @click="toggleTestDetails(index)">
          <div class="test-status" :class="test.status">
            <span v-if="test.status === 'passed'">✓</span>
            <span v-else>✗</span>
          </div>
          <div class="test-name">{{ test.name }}</div>
          <div class="test-endpoint">{{ test.endpoint }}</div>
          <div class="test-time">{{ test.time }}ms</div>
          <div class="test-date">{{ test.date }}</div>
          <div class="test-expand">
            <span>{{ test.expanded ? '▼' : '▶' }}</span>
          </div>
        </div>

        <div class="test-details" v-if="test.expanded">
          <div class="test-details-section">
            <h4>Request</h4>
            <div class="details-content">
              <p><strong>Method:</strong> {{ test.request.method }}</p>
              <p><strong>URL:</strong> {{ test.request.url }}</p>
              <div v-if="test.request.headers.length > 0">
                <p><strong>Headers:</strong></p>
                <ul>
                  <li v-for="(header, i) in test.request.headers" :key="i">
                    {{ header.key }}: {{ header.value }}
                  </li>
                </ul>
              </div>
              <div v-if="test.request.body">
                <p><strong>Body:</strong></p>
                <pre>{{ JSON.stringify(test.request.body, null, 2) }}</pre>
              </div>
            </div>
          </div>

          <div class="test-details-section">
            <h4>Response</h4>
            <div class="details-content">
              <p><strong>Status:</strong> {{ test.response.status }} {{ test.response.statusText }}</p>
              <p><strong>Time:</strong> {{ test.response.time }}ms</p>
              <div v-if="test.response.headers.length > 0">
                <p><strong>Headers:</strong></p>
                <ul>
                  <li v-for="(header, i) in test.response.headers" :key="i">
                    {{ header.key }}: {{ header.value }}
                  </li>
                </ul>
              </div>
              <div v-if="test.response.body">
                <p><strong>Body:</strong></p>
                <pre>{{ JSON.stringify(test.response.body, null, 2) }}</pre>
              </div>
            </div>
          </div>

          <div class="test-details-section">
            <h4>Test Results</h4>
            <div class="details-content">
              <div class="assertion"
                   v-for="(assertion, i) in test.assertions"
                   :key="i"
                   :class="assertion.passed ? 'passed' : 'failed'">
                <span class="assertion-status">
                  {{ assertion.passed ? '✓' : '✗' }}
                </span>
                <span class="assertion-name">{{ assertion.name }}</span>
                <span class="assertion-message" v-if="!assertion.passed">
                  {{ assertion.message }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="pagination">
      <button :disabled="currentPage === 1" @click="currentPage--">Previous</button>
      <span>Page {{ currentPage }} of {{ totalPages }}</span>
      <button :disabled="currentPage === totalPages" @click="currentPage++">Next</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'TestResults',
  data() {
    return {
      filters: {
        endpoint: '',
        status: '',
        date: ''
      },
      currentPage: 1,
      testsPerPage: 5,
      tests: []
    };
  },
  computed: {
    filteredTests() {
      let filtered = this.tests;

      if (this.filters.endpoint) {
        filtered = filtered.filter(test => test.endpoint === this.filters.endpoint);
      }

      if (this.filters.status) {
        filtered = filtered.filter(test => test.status === this.filters.status);
      }

      if (this.filters.date) {
        filtered = filtered.filter(test => test.date === this.filters.date);
      }

      const start = (this.currentPage - 1) * this.testsPerPage;
      const end = start + this.testsPerPage;
      return filtered.slice(start, end);
    },
    totalPages() {
      const filtered = this.tests.filter(test => {
        if (this.filters.endpoint && test.endpoint !== this.filters.endpoint) return false;
        if (this.filters.status && test.status !== this.filters.status) return false;
        if (this.filters.date && test.date !== this.filters.date) return false;
        return true;
      });

      return Math.ceil(filtered.length / this.testsPerPage);
    },
    testSummary() {
      const total = this.tests.length;
      const passed = this.tests.filter(test => test.status === 'passed').length;
      const failed = total - passed;
      const successRate = total > 0 ? Math.round((passed / total) * 100) : 0;

      return {
        total,
        passed,
        failed,
        successRate
      };
    }
  },
  methods: {
    toggleTestDetails(index) {
      this.$set(this.filteredTests[index], 'expanded', !this.filteredTests[index].expanded);
    },
    applyFilters() {
      this.currentPage = 1; // Сбрасываем страницу при применении фильтров
    },
    loadTests() {
      const storedTests = JSON.parse(localStorage.getItem('testResults') || '[]');
      this.tests = storedTests.map(test => ({
        ...test,
        expanded: false // Гарантируем, что все тесты изначально свернуты
      }));
    }
  },
  mounted() {
    this.loadTests();
    // Обновляем тесты при изменении localStorage (например, после нового запроса)
    window.addEventListener('storage', this.loadTests);
  },
  beforeDestroy() {
    window.removeEventListener('storage', this.loadTests);
  }
};
</script>

<style scoped>
.test-results {
  padding: 24px;
}

.test-summary {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.summary-card {
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

.stat-value.success {
  color: #4caf50;
}

.stat-value.error {
  color: #f44336;
}

.results-filters {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  flex-wrap: wrap;
  align-items: flex-end;
}

.filter-item {
  display: flex;
  flex-direction: column;
}

.filter-item label {
  margin-bottom: 8px;
  font-size: 14px;
}

.filter-apply {
  height: 36px;
}

.test-list {
  margin-bottom: 24px;
}

.test-item {
  border: 1px solid var(--border-color);
  border-radius: 4px;
  margin-bottom: 8px;
  overflow: hidden;
}

.test-header {
  display: flex;
  padding: 12px;
  cursor: pointer;
  background-color: var(--secondary-color);
  align-items: center;
}

.test-header:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.test-status {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
}

.test-status.passed {
  color: #4caf50;
}

.test-status.failed {
  color: #f44336;
}

.test-name {
  flex: 1;
  font-weight: bold;
}

.test-endpoint {
  width: 150px;
  color: var(--accent-color);
}

.test-time {
  width: 100px;
  text-align: right;
}

.test-date {
  width: 100px;
  text-align: right;
  margin-right: 12px;
}

.test-expand {
  width: 24px;
  text-align: center;
}

.test-details {
  padding: 16px;
  background-color: var(--input-bg-color);
  border-top: 1px solid var(--border-color);
}

.test-details-section {
  margin-bottom: 16px;
}

.test-details-section h4 {
  margin-bottom: 8px;
  color: var(--accent-color);
}

.details-content {
  padding-left: 16px;
}

.details-content p {
  margin-bottom: 8px;
}

.details-content pre {
  background-color: rgba(0, 0, 0, 0.2);
  padding: 8px;
  border-radius: 4px;
  overflow-x: auto;
  font-family: monospace;
}

.assertion {
  display: flex;
  align-items: center;
  padding: 4px 0;
}

.assertion-status {
  width: 24px;
  margin-right: 8px;
}

.assertion.passed .assertion-status {
  color: #4caf50;
}

.assertion.failed .assertion-status {
  color: #f44336;
}

.assertion-name {
  flex: 1;
}

.assertion-message {
  color: #f44336;
  margin-left: 16px;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
}
</style>