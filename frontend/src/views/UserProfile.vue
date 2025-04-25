<template>
  <div class="profile-container">
    <h2>Your Profile</h2>

    <div class="profile-card">
      <div class="form-group">
        <label for="name">Full Name</label>
        <input
            type="text"
            id="name"
            v-model="form.name"
            placeholder="Your name"
        />
      </div>

      <div class="form-group">
        <label for="email">Email</label>
        <input
            type="email"
            id="email"
            v-model="form.email"
            placeholder="Your email"
            disabled
        />
        <p class="help-text">Email cannot be changed</p>
      </div>

      <div class="form-divider"></div>

      <h3>Change Password</h3>

      <div class="form-group">
        <label for="currentPassword">Current Password</label>
        <input
            type="password"
            id="currentPassword"
            v-model="form.currentPassword"
            placeholder="Current password"
        />
      </div>

      <div class="form-group">
        <label for="newPassword">New Password</label>
        <input
            type="password"
            id="newPassword"
            v-model="form.newPassword"
            placeholder="New password"
        />
      </div>

      <div class="form-group">
        <label for="confirmPassword">Confirm New Password</label>
        <input
            type="password"
            id="confirmPassword"
            v-model="form.confirmPassword"
            placeholder="Confirm new password"
        />
      </div>

      <div class="form-group">
        <button class="accent-button" @click="updateProfile">Update Profile</button>
      </div>

      <div class="success-message" v-if="updateSuccess">
        Profile updated successfully!
      </div>
    </div>

    <div class="api-keys-section">
      <h3>API Keys</h3>
      <p>Generate API keys to use our service programmatically</p>

      <div class="api-key-list">
        <div class="api-key-item" v-for="(key, index) in apiKeys" :key="index">
          <div class="key-info">
            <div class="key-name">{{ key.name }}</div>
            <div class="key-value">
              <span v-if="key.visible">{{ key.value }}</span>
              <span v-else>••••••••••••••••</span>
            </div>
            <div class="key-created">Created: {{ key.created }}</div>
          </div>
          <div class="key-actions">
            <button @click="toggleKeyVisibility(index)">
              {{ key.visible ? 'Hide' : 'Show' }}
            </button>
            <button class="delete-button" @click="deleteKey(index)">Delete</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import axios from 'axios'

const form = reactive({
  name: '',
  email: '',
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const updateSuccess = ref(false)
const apiKeys = ref([])

onMounted(async () => {
  await loadProfile()
  await loadApiKeys()
})

async function loadProfile() {
  try {
    const { data } = await axios.get('/api/profile')
    form.name = data.name
    form.email = data.email
  } catch (err) {
    console.error('Failed to load profile:', err)
  }
}

async function loadApiKeys() {
  try {
    const { data } = await axios.get('/api/keys')
    // Ожидаем, что сервер вернёт массив { id, name, value, created }
    apiKeys.value = data.map(key => ({
      ...key,
      visible: false
    }))
  } catch (err) {
    console.error('Failed to load API keys:', err)
  }
}

async function updateProfile() {
  if (form.newPassword !== form.confirmPassword) {
    alert('Новые пароли не совпадают')
    return
  }

  try {
    await axios.put('/api/profile', {
      name: form.name,
      currentPassword: form.currentPassword,
      newPassword: form.newPassword
    })
    updateSuccess.value = true

    // Очистим поля пароля
    form.currentPassword = ''
    form.newPassword = ''
    form.confirmPassword = ''

    setTimeout(() => {
      updateSuccess.value = false
    }, 3000)
  } catch (err) {
    console.error('Update failed:', err)
    alert(err.response?.data?.message || 'Ошибка при обновлении профиля')
  }
}

function toggleKeyVisibility(index) {
  apiKeys.value[index].visible = !apiKeys.value[index].visible
}

async function deleteKey(index) {
  const key = apiKeys.value[index]
  if (!confirm(`Удалить ключ API «${key.name}»?`)) return

  try {
    await axios.delete(`/api/keys/${key.id}`)
    apiKeys.value.splice(index, 1)
  } catch (err) {
    console.error('Delete failed:', err)
    alert('Ошибка при удалении ключа')
  }
}
</script>

<style scoped>

</style>