<template>
  <div class="request-builder">
    <div class="request-form">
      <div class="request-info">
        <div class="method-select">
          <select v-model="request.method">
            <option>GET</option>
            <option>POST</option>
            <option>PUT</option>
            <option>DELETE</option>
            <option>PATCH</option>
          </select>
        </div>
        <div class="url-input">
          <input
              type="text"
              v-model="request.url"
              placeholder="Enter request URL (e.g., /api/v2/tasks)"
              @keyup.enter="sendRequest"
          />
        </div>
        <div class="send-button">
          <button class="accent-button" @click="sendRequest" :disabled="isLoading">
            {{ isLoading ? 'Sending...' : 'Send' }}
          </button>
        </div>
      </div>

      <div class="request-details">
        <div class="tabs">
          <div
              v-for="tab in tabs"
              :key="tab.id"
              :class="['tab', { active: activeTab === tab.id }]"
              @click="activeTab = tab.id"
          >
            {{ tab.name }}
          </div>
        </div>

        <div class="tab-content">
          <!-- Параметры -->
          <div v-if="activeTab === 'params'" class="params-container">
            <div class="param-row header">
              <div class="check">Use</div>
              <div class="key">Key</div>
              <div class="value">Value</div>
              <div class="description">Description</div>
              <div class="actions"></div>
            </div>
            <div
                v-for="(param, index) in request.params"
                :key="index"
                class="param-row"
            >
              <div class="check">
                <input type="checkbox" v-model="param.enabled" />
              </div>
              <div class="key">
                <input type="text" v-model="param.key" placeholder="Parameter name" />
              </div>
              <div class="value">
                <input type="text" v-model="param.value" placeholder="Parameter value" />
              </div>
              <div class="description">
                <input type="text" v-model="param.description" placeholder="Parameter description" />
              </div>
              <div class="actions">
                <button @click="removeParam(index)">×</button>
              </div>
            </div>
            <button @click="addParam" class="add-param">+ Add Parameter</button>
          </div>

          <!-- Заголовки -->
          <div v-if="activeTab === 'headers'" class="headers-container">
            <div class="param-row header">
              <div class="check">Use</div>
              <div class="key">Header</div>
              <div class="value">Value</div>
              <div class="description">Description</div>
              <div class="actions"></div>
            </div>
            <div
                v-for="(header, index) in request.headers"
                :key="index"
                class="param-row"
            >
              <div class="check">
                <input type="checkbox" v-model="header.enabled" />
              </div>
              <div class="key">
                <input type="text" v-model="header.key" placeholder="Header name" />
              </div>
              <div class="value">
                <input type="text" v-model="header.value" placeholder="Header value" />
              </div>
              <div class="description">
                <input type="text" v-model="header.description" placeholder="Header description" />
              </div>
              <div class="actions">
                <button @click="removeHeader(index)">×</button>
              </div>
            </div>
            <button @click="addHeader" class="add-param">+ Add Header</button>
          </div>

          <!-- Тело запроса -->
          <div v-if="activeTab === 'body'" class="body-container">
            <div class="body-type-select">
              <select v-model="request.bodyType">
                <option value="none">None</option>
                <option value="raw">Raw</option>
              </select>
              <select v-if="request.bodyType === 'raw'" v-model="request.bodyFormat">
                <option value="json">JSON</option>
              </select>
            </div>
            <div v-if="request.bodyType === 'raw'" class="raw-editor">
              <textarea
                  v-model="request.rawBody"
                  placeholder='Enter JSON content, e.g., {"type": "example_task", "template_id": "template_123", "template": "Sample template", "amount": 10}'
                  rows="10"
              ></textarea>
            </div>
          </div>

          <!-- Тесты -->
          <div v-if="activeTab === 'tests'" class="tests-container">
            <div class="test-options">
              <div class="test-option">
                <label>
                  <input type="checkbox" v-model="testSettings.validateResponse" />
                  Validate Response Schema
                </label>
              </div>
              <div class="test-option">
                <label>
                  <input type="checkbox" v-model="testSettings.checkStatusCode" />
                  Check Status Code
                </label>
                <input
                    v-if="testSettings.checkStatusCode"
                    type="text"
                    v-model="testSettings.expectedStatusCode"
                    placeholder="Expected Status (e.g. 200)"
                />
              </div>
              <div class="test-option">
                <label>
                  <input type="checkbox" v-model="testSettings.checkResponseTime" />
                  Check Response Time
                </label>
                <input
                    v-if="testSettings.checkResponseTime"
                    type="text"
                    v-model="testSettings.maxResponseTime"
                    placeholder="Max Response Time (ms)"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="response-section" v-if="response">
      <div class="response-header">
        <h2>Response</h2>
        <div class="response-meta">
          <span class="status" :class="responseStatusClass">
            Status: {{ response.status }} {{ response.statusText }}
          </span>
          <span class="time">Time: {{ response.time }}ms</span>
          <span class="size">Size: {{ response.size }}</span>
        </div>
      </div>

      <div class="response-body">
        <pre>{{ JSON.stringify(response.body, null, 2) }}</pre>
      </div>

      <div class="test-generation-actions">
        <button class="accent-button" @click="generateTestData">Generate Test Data</button>
      </div>
    </div>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import authStore from '../store/auth';
import Ajv from 'ajv';

const taskSchema = {
  type: 'object',
  properties: {
    id: { type: 'integer' },
    task_id: { type: 'string' },
    user_id: { type: 'string' },
    type: { type: 'string' },
    template_id: { type: 'string' },
    template: { type: 'string' },
    amount: { type: 'integer' }
  },
  required: ['id', 'task_id', 'user_id', 'type'],
  additionalProperties: false
};

export default {
  name: 'RequestBuilder',
  data() {
    return {
      activeTab: 'body',
      tabs: [
        { id: 'params', name: 'Params' },
        { id: 'headers', name: 'Headers' },
        { id: 'body', name: 'Body' },
        { id: 'tests', name: 'Tests' }
      ],
      request: {
        method: 'POST',
        url: '/api/v2/tasks',
        params: [],
        headers: [
          { key: 'Content-Type', value: 'application/json', description: 'Content type', enabled: true },
          { key: 'Authorization', value: '', description: 'Bearer token', enabled: true }
        ],
        bodyType: 'raw',
        bodyFormat: 'json',
        bodyParams: [],
        rawBody: JSON.stringify(
            {
              type: '',
              template_id: '',
              template: '',
              amount: 0
            },
            null,
            2
        )
      },
      testSettings: {
        validateResponse: true,
        checkStatusCode: true,
        expectedStatusCode: '200',
        checkResponseTime: false,
        maxResponseTime: '1000',
        checkHeaderPresence: false,
        expectedHeader: '',
        testCasesCount: 5,
        includeEdgeCases: true,
        generateNegativeTests: true
      },
      response: null,
      error: null,
      isLoading: false
    };
  },
  computed: {
    responseStatusClass() {
      if (!this.response) return '';
      const status = this.response.status;
      if (status >= 200 && status < 300) return 'success';
      if (status >= 400 && status < 500) return 'client-error';
      if (status >= 500) return 'server-error';
      return '';
    }
  },
  methods: {
    addParam() {
      this.request.params.push({ key: '', value: '', description: '', enabled: true });
    },
    removeParam(index) {
      this.request.params.splice(index, 1);
    },
    addHeader() {
      this.request.headers.push({ key: '', value: '', description: '', enabled: true });
    },
    removeHeader(index) {
      this.request.headers.splice(index, 1);
    },
    addBodyParam() {
      this.request.bodyParams.push({ key: '', value: '', description: '', enabled: true, type: 'text' });
    },
    removeBodyParam(index) {
      this.request.bodyParams.splice(index, 1);
    },
    async sendRequest() {
      this.response = null;
      this.error = null;
      this.isLoading = true;

      if (!authStore.state.isAuthenticated || !authStore.state.token) {
        this.error = 'Please log in to send requests to task-service';
        this.isLoading = false;
        return;
      }

      const headers = {};
      this.request.headers.forEach(header => {
        if (header.enabled && header.key) {
          headers[header.key] = header.value;
        }
      });
      headers['Authorization'] = `Bearer ${authStore.state.token}`;

      const params = {};
      this.request.params.forEach(param => {
        if (param.enabled && param.key) {
          params[param.key] = param.value;
        }
      });

      let data = null;
      if (this.request.bodyType === 'raw' && this.request.rawBody) {
        try {
          data = JSON.parse(this.request.rawBody);
        } catch (e) {
          this.error = 'Invalid JSON in request body';
          this.isLoading = false;
          return;
        }
      }

      try {
        const startTime = performance.now();
        const response = await axios({
          method: this.request.method,
          url: this.request.url,
          headers,
          params,
          data
        });

        const endTime = performance.now();
        const responseSize = JSON.stringify(response.data).length;

        this.response = {
          status: response.status,
          statusText: response.statusText,
          time: Math.round(endTime - startTime),
          size: `${(responseSize / 1024).toFixed(2)} KB`,
          body: response.data,
          headers: Object.entries(response.headers).map(([key, value]) => ({ key, value }))
        };

        // Формируем assertions
        const assertions = [];

        if (this.testSettings.checkStatusCode) {
          const passed = response.status === parseInt(this.testSettings.expectedStatusCode);
          assertions.push({
            name: `Status code is ${this.testSettings.expectedStatusCode}`,
            passed,
            message: passed ? '' : `Expected ${this.testSettings.expectedStatusCode}, got ${response.status}`
          });
        }

        if (this.testSettings.validateResponse) {
          const ajv = new Ajv();
          const validate = ajv.compile(taskSchema);
          const valid = validate(response.data);
          assertions.push({
            name: 'Response has valid schema',
            passed: valid,
            message: valid ? '' : ajv.errorsText(validate.errors)
          });
        }

        if (this.testSettings.checkResponseTime) {
          const maxTime = parseInt(this.testSettings.maxResponseTime);
          const passed = this.response.time <= maxTime;
          assertions.push({
            name: `Response time is less than ${maxTime}ms`,
            passed,
            message: passed ? '' : `Response time ${this.response.time}ms exceeds ${maxTime}ms`
          });
        }

        // Сохраняем результат теста в localStorage
        const testResult = {
          id: Date.now(), // Уникальный ID на основе времени
          name: `Test ${this.request.method} ${this.request.url}`,
          endpoint: `${this.request.method} ${this.request.url}`,
          status: assertions.every(a => a.passed) ? 'passed' : 'failed',
          time: this.response.time,
          date: new Date().toISOString().split('T')[0],
          expanded: false,
          request: {
            method: this.request.method,
            url: this.request.url,
            headers: this.request.headers.filter(h => h.enabled),
            body: data
          },
          response: this.response,
          assertions
        };

        const storedTests = JSON.parse(localStorage.getItem('testResults') || '[]');
        storedTests.push(testResult);
        localStorage.setItem('testResults', JSON.stringify(storedTests));
      } catch (error) {
        this.error = error.response?.data?.error || 'Request failed';
        this.response = {
          status: error.response?.status || 500,
          statusText: error.response?.statusText || 'Error',
          time: 0,
          size: '0 KB',
          body: error.response?.data || { error: 'Request failed' },
          headers: []
        };

        // Сохраняем результат неуспешного теста
        const testResult = {
          id: Date.now(),
          name: `Test ${this.request.method} ${this.request.url}`,
          endpoint: `${this.request.method} ${this.request.url}`,
          status: 'failed',
          time: 0,
          date: new Date().toISOString().split('T')[0],
          expanded: false,
          request: {
            method: this.request.method,
            url: this.request.url,
            headers: this.request.headers.filter(h => h.enabled),
            body: data
          },
          response: this.response,
          assertions: [
            {
              name: `Status code is ${this.testSettings.expectedStatusCode}`,
              passed: false,
              message: `Request failed: ${this.error}`
            }
          ]
        };

        const storedTests = JSON.parse(localStorage.getItem('testResults') || '[]');
        storedTests.push(testResult);
        localStorage.setItem('testResults', JSON.stringify(storedTests));
      } finally {
        this.isLoading = false;
      }
    },
    generateTestData() {
      alert('Test data generation not implemented yet');
    }
  },
  mounted() {
    if (authStore.state.token) {
      const authHeader = this.request.headers.find(h => h.key === 'Authorization');
      if (authHeader) {
        authHeader.value = `Bearer ${authStore.state.token}`;
      }
    }
  }
};
</script>

<style scoped>
.request-builder {
  padding: 24px;
}

.request-form {
  margin-bottom: 24px;
}

.request-info {
  display: flex;
  margin-bottom: 16px;
}

.method-select {
  width: 100px;
  margin-right: 8px;
}

.url-input {
  flex: 1;
  margin-right: 8px;
}

.tabs {
  display: flex;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 16px;
}

.tab {
  padding: 8px 16px;
  cursor: pointer;
  border-bottom: 2px solid transparent;
}

.tab.active {
  border-bottom: 2px solid var(--accent-color);
  color: var(--accent-color);
}

.param-row {
  display: flex;
  margin-bottom: 8px;
}

.param-row.header {
  font-weight: bold;
  color: var(--accent-color);
  margin-bottom: 8px;
}

.check {
  width: 40px;
  display: flex;
  align-items: center;
}

.key, .value {
  flex: 1;
  margin-right: 8px;
}

.description {
  flex: 2;
  margin-right: 8px;
}

.actions {
  width: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.actions button {
  background: none;
  border: none;
  font-size: 18px;
  color: var(--text-color);
  cursor: pointer;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.add-param {
  margin-top: 8px;
  background: none;
  border: 1px dashed var(--border-color);
  width: 100%;
  padding: 8px;
  color: var(--accent-color);
}

.raw-editor textarea {
  width: 100%;
  font-family: monospace;
  background-color: var(--input-bg-color);
  color: var(--text-color);
  border: 1px solid var(--border-color);
  padding: 8px;
}

.body-type-select {
  display: flex;
  margin-bottom: 16px;
  gap: 8px;
}

.body-type-select select {
  width: auto;
}

.test-options {
  margin-bottom: 24px;
}

.test-option {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
}

.test-option label {
  display: flex;
  align-items: center;
  margin-right: 16px;
}

.test-option input[type="checkbox"] {
  margin-right: 8px;
  width: auto;
}

.test-option input[type="text"] {
  width: 200px;
}

.test-generation {
  border-top: 1px solid var(--border-color);
  padding-top: 16px;
}

.test-generation h3 {
  margin-bottom: 16px;
}

.test-generation-options {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 16px;
}

.test-generation-option {
  display: flex;
  flex-direction: column;
}

.test-generation-option label {
  margin-bottom: 8px;
}

.response-section {
  background-color: var(--secondary-color);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 16px;
}

.response-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.response-meta {
  display: flex;
  gap: 16px;
}

.status {
  padding: 4px 8px;
  border-radius: 4px;
}

.status.success {
  background-color: rgba(76, 175, 80, 0.2);
  color: #4caf50;
}

.status.client-error {
  background-color: rgba(255, 152, 0, 0.2);
  color: #ff9800;
}

.status.server-error {
  background-color: rgba(244, 67, 54, 0.2);
  color: #f44336;
}

.response-body {
  background-color: var(--input-bg-color);
  padding: 16px;
  border-radius: 4px;
  margin-bottom: 16px;
  overflow-x: auto;
}

.response-body pre {
  font-family: monospace;
  white-space: pre-wrap;
}

.test-generation-actions {
  display: flex;
  justify-content: flex-end;
}
</style>