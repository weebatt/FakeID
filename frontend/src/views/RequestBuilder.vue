// views/RequestBuilder.vue
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
              placeholder="Enter request URL"
              @keyup.enter="sendRequest"
          />
        </div>
        <div class="send-button">
          <button class="accent-button" @click="sendRequest">Send</button>
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
                <option value="form-data">Form Data</option>
                <option value="x-www-form-urlencoded">x-www-form-urlencoded</option>
                <option value="raw">Raw</option>
                <option value="binary">Binary</option>
              </select>

              <select v-if="request.bodyType === 'raw'" v-model="request.bodyFormat">
                <option value="json">JSON</option>
                <option value="text">Text</option>
                <option value="xml">XML</option>
                <option value="html">HTML</option>
              </select>
            </div>

            <div v-if="request.bodyType === 'form-data' || request.bodyType === 'x-www-form-urlencoded'">
              <div class="param-row header">
                <div class="check">Use</div>
                <div class="key">Key</div>
                <div class="value">Value</div>
                <div class="description">Description</div>
                <div class="actions"></div>
              </div>
              <div
                  v-for="(param, index) in request.bodyParams"
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
                  <input
                      v-if="request.bodyType === 'form-data' && param.type === 'file'"
                      type="file"
                  />
                  <input
                      v-else
                      type="text"
                      v-model="param.value"
                      placeholder="Parameter value"
                  />
                </div>
                <div class="description">
                  <input type="text" v-model="param.description" placeholder="Parameter description" />
                </div>
                <div class="actions">
                  <button @click="removeBodyParam(index)">×</button>
                </div>
              </div>
              <button @click="addBodyParam" class="add-param">+ Add Parameter</button>
            </div>

            <div v-else-if="request.bodyType === 'raw'" class="raw-editor">
              <textarea
                  v-model="request.rawBody"
                  :placeholder="`Enter ${request.bodyFormat.toUpperCase()} content`"
                  rows="10"
              ></textarea>
            </div>

            <div v-else-if="request.bodyType === 'binary'" class="binary-editor">
              <input type="file" />
            </div>
          </div>

          <!-- Настройки тестов -->
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
              <div class="test-option">
                <label>
                  <input type="checkbox" v-model="testSettings.checkHeaderPresence" />
                  Check Header Presence
                </label>
                <input
                    v-if="testSettings.checkHeaderPresence"
                    type="text"
                    v-model="testSettings.expectedHeader"
                    placeholder="Header Name"
                />
              </div>
            </div>

            <div class="test-generation">
              <h3>Test Data Generation</h3>
              <div class="test-generation-options">
                <div class="test-generation-option">
                  <label>Number of Test Cases</label>
                  <input type="number" v-model="testSettings.testCasesCount" min="1" max="100" />
                </div>
                <div class="test-generation-option">
                  <label>Include Edge Cases</label>
                  <input type="checkbox" v-model="testSettings.includeEdgeCases" />
                </div>
                <div class="test-generation-option">
                  <label>Generate Negative Tests</label>
                  <input type="checkbox" v-model="testSettings.generateNegativeTests" />
                </div>
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
  </div>
</template>

<script>
export default {
  name: 'RequestBuilder',
  data() {
    return {
      activeTab: 'params',
      tabs: [
        { id: 'params', name: 'Params' },
        { id: 'headers', name: 'Headers' },
        { id: 'body', name: 'Body' },
        { id: 'tests', name: 'Tests' }
      ],
      request: {
        method: 'GET',
        url: '',
        params: [],
        headers: [
          { key: 'Content-Type', value: 'application/json', description: '', enabled: true }
        ],
        bodyType: 'none',
        bodyFormat: 'json',
        bodyParams: [],
        rawBody: ''
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
      response: null
    }
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
    sendRequest() {
      // В реальном приложении здесь был бы запрос к API
      // Имитация ответа для демонстрации
      setTimeout(() => {
        this.response = {
          status: 200,
          statusText: 'OK',
          time: 235,
          size: '1.2 KB',
          body: {
            success: true,
            data: {
              id: 1,
              name: 'Test User',
              email: 'test@example.com',
              created_at: '2023-01-01T00:00:00Z'
            }
          }
        };
      }, 500);
    },
    generateTestData() {
      // Логика генерации тестовых данных
      alert('Test data will be generated and packaged as ZIP');
      // Здесь можно добавить логику для скачивания ZIP-архива
    }
  }
}
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