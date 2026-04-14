<template>
  <div class="course-workspace">
    <header class="card workspace-header">
      <h2>AI 模型管理</h2>
      <p class="subtle">管理问答、向量化与重排模型配置</p>
    </header>
    <div v-if="error" class="alert error">{{ error }}</div>
    <div v-if="success" class="alert success">{{ success }}</div>
    <section class="card stack">
      <div class="list-toolbar-actions">
        <input v-model="form.name" placeholder="名称" />
        <select v-model="form.model_type">
          <option value="qa">qa</option>
          <option value="embedding">embedding</option>
          <option value="rerank">rerank</option>
        </select>
        <input v-model="form.api_base_url" placeholder="API Base URL" />
        <input v-model="form.model_id" placeholder="Model ID" />
        <input v-model="form.api_key" type="password" placeholder="API Key（编辑时留空表示沿用）" />
        <button class="primary" @click="handleSubmit">{{ editingId ? '保存修改' : '新增' }}</button>
        <button v-if="editingId" @click="resetForm">取消编辑</button>
        <button :disabled="testing" @click="handleTest">{{ testing ? '测试中...' : '测试连通性' }}</button>
      </div>
      <div class="course-list">
        <article v-for="m in items" :key="m.id" class="course-item">
          <div>
            <h4>{{ m.name }}</h4>
            <p class="mono">{{ m.model_type }} · {{ m.model_id }}</p>
            <p class="subtle">{{ m.api_base_url }}</p>
            <p class="subtle">API Key：{{ m.has_api_key ? '已保存' : '未保存' }}</p>
          </div>
          <div class="action-row">
            <button @click="startEdit(m)">编辑</button>
            <button class="danger" @click="handleDelete(m.id)">删除</button>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import {
  listAIModels,
  createAIModel,
  updateAIModel,
  deleteAIModel,
  testAIModelConnection,
  type AIModelItem,
} from '@/services/aiModels'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const items = ref<AIModelItem[]>([])
const error = ref('')
const success = ref('')
const testing = ref(false)
const editingId = ref('')
const form = reactive({
  name: '',
  model_type: 'qa' as 'qa' | 'embedding' | 'rerank',
  api_base_url: '',
  model_id: '',
  api_key: '',
})

async function load() {
  // 加载当前教师下的全部模型配置。
  error.value = ''
  try {
    const data = await listAIModels(auth.token ?? '')
    items.value = data.items || []
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载模型失败'
  }
}

function resetForm() {
  editingId.value = ''
  form.name = ''
  form.model_type = 'qa'
  form.api_base_url = ''
  form.model_id = ''
  form.api_key = ''
}

function startEdit(item: AIModelItem) {
  editingId.value = item.id
  form.name = item.name
  form.model_type = item.model_type
  form.api_base_url = item.api_base_url
  form.model_id = item.model_id
  form.api_key = ''
}

// 统一处理新增/编辑，避免表单状态重复分叉。
async function handleSubmit() {
  error.value = ''
  success.value = ''
  try {
    if (editingId.value) {
      await updateAIModel(auth.token ?? '', editingId.value, {
        name: form.name,
        api_base_url: form.api_base_url,
        model_id: form.model_id,
        api_key: form.api_key,
      })
      success.value = '模型更新成功'
    } else {
      await createAIModel(auth.token ?? '', form)
      success.value = '模型创建成功'
    }
    resetForm()
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '提交模型失败'
  }
}

// 连接测试支持新增场景和编辑场景（可复用已有密钥）。
async function handleTest() {
  error.value = ''
  success.value = ''
  testing.value = true
  try {
    const out = await testAIModelConnection(auth.token ?? '', {
      model_type: form.model_type,
      api_base_url: form.api_base_url,
      model_id: form.model_id,
      api_key: form.api_key || undefined,
      existing_model_id: editingId.value || undefined,
    })
    success.value = out.message || (out.ok ? '连通性测试通过' : '连通性测试失败')
  } catch (err) {
    error.value = err instanceof Error ? err.message : '连通性测试失败'
  } finally {
    testing.value = false
  }
}

async function handleDelete(id: string) {
  // 删除指定模型并重新拉取列表。
  error.value = ''
  success.value = ''
  try {
    await deleteAIModel(auth.token ?? '', id)
    success.value = '模型删除成功'
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除模型失败'
  }
}

onMounted(() => void load())
</script>

<style scoped>
.list-toolbar-actions {
  display: grid;
  grid-template-columns: repeat(3, minmax(180px, 1fr));
  gap: 10px;
}

.list-toolbar-actions > input,
.list-toolbar-actions > select,
.list-toolbar-actions > button {
  width: 100%;
}

.course-item {
  align-items: center;
}

.course-item h4 {
  margin-bottom: 4px;
  font-size: 16px;
}

.course-item .action-row {
  align-items: center;
}

.course-item .subtle {
  margin: 2px 0;
}

.success {
  background: #ecfdf5;
  border-color: #86efac;
  color: #166534;
}

@media (max-width: 1200px) {
  .list-toolbar-actions {
    grid-template-columns: repeat(2, minmax(180px, 1fr));
  }
}

@media (max-width: 760px) {
  .list-toolbar-actions {
    grid-template-columns: 1fr;
  }
}
</style>
