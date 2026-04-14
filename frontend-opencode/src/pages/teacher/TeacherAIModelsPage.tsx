import { useEffect, useState } from 'react'
import {
  getModels,
  createModel,
  updateModel,
  deleteModel,
  testConnection,
} from '@/services/aiModels'
import type { AIModelListItem, AIModelType } from '@/types'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'

const MODEL_TYPES: { value: AIModelType; label: string; description: string }[] = [
  { value: 'qa', label: '问答', description: '用于对话问答的模型' },
  { value: 'embedding', label: '嵌入', description: '用于向量嵌入的模型' },
  { value: 'rerank', label: '重排', description: '用于结果重排序的模型' },
]

/**
 * TeacherAIModelsPage - AI模型管理页面
 * 管理教师的问答/嵌入/重排模型配置
 */
export function TeacherAIModelsPage() {
  const [models, setModels] = useState<AIModelListItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<'create' | 'edit'>('create')
  const [editingModel, setEditingModel] = useState<AIModelListItem | null>(null)
  const [formData, setFormData] = useState({
    name: '',
    model_type: 'embedding' as AIModelType,
    api_base_url: '',
    model_id: '',
    api_key: '',
  })
  const [formErrors, setFormErrors] = useState<Record<string, string>>({})

  const [testLoading, setTestLoading] = useState(false)
  const [testResult, setTestResult] = useState<{ ok: boolean; text: string } | null>(null)

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [modelToDelete, setModelToDelete] = useState<AIModelListItem | null>(null)

  /**
   * 加载模型列表
   */
  const loadModels = async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await getModels()
      setModels(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载模型列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadModels()
  }, [])

  /**
   * 打开创建模型对话框
   */
  const handleOpenCreateDialog = () => {
    setDialogMode('create')
    setEditingModel(null)
    setFormData({
      name: '',
      model_type: 'embedding',
      api_base_url: 'https://api.openai.com/v1/embeddings',
      model_id: 'text-embedding-3-small',
      api_key: '',
    })
    setFormErrors({})
    setTestResult(null)
    setDialogOpen(true)
  }

  /**
   * 打开编辑模型对话框
   */
  const handleOpenEditDialog = (model: AIModelListItem) => {
    setDialogMode('edit')
    setEditingModel(model)
    setFormData({
      name: model.name,
      model_type: model.model_type,
      api_base_url: model.api_base_url,
      model_id: model.model_id,
      api_key: '',
    })
    setFormErrors({})
    setTestResult(null)
    setDialogOpen(true)
  }

  /**
   * 表单验证
   */
  const validateForm = (): boolean => {
    const errors: Record<string, string> = {}

    if (!formData.name.trim()) {
      errors.name = '模型名称不能为空'
    }

    if (!formData.api_base_url.trim()) {
      errors.api_base_url = 'API 地址不能为空'
    }

    if (!formData.model_id.trim()) {
      errors.model_id = '模型 ID 不能为空'
    }

    if (dialogMode === 'create' && !formData.api_key.trim()) {
      errors.api_key = '创建模型时必须提供 API Key'
    }

    setFormErrors(errors)
    return Object.keys(errors).length === 0
  }

  /**
   * 测试连接是否可用
   */
  const canTestConnection =
    formData.api_base_url.trim() !== '' &&
    formData.model_id.trim() !== '' &&
    (formData.api_key.trim() !== '' || (editingModel?.has_api_key ?? false))

  /**
   * 执行连接测试
   */
  const handleTestConnection = async () => {
    if (!canTestConnection) return

    setTestLoading(true)
    setTestResult(null)
    try {
      const result = await testConnection({
        model_type: formData.model_type,
        api_base_url: formData.api_base_url.trim(),
        model_id: formData.model_id.trim(),
        api_key: formData.api_key.trim() || undefined,
        existing_model_id:
          editingModel && !formData.api_key.trim() && editingModel.has_api_key
            ? Number(editingModel.id)
            : undefined,
      })
      setTestResult({
        ok: result.ok,
        text: result.message + (result.http_status ? ` (HTTP ${result.http_status})` : ''),
      })
    } catch (err) {
      setTestResult({
        ok: false,
        text: err instanceof Error ? err.message : '测试失败',
      })
    } finally {
      setTestLoading(false)
    }
  }

  /**
   * 提交表单
   */
  const handleSubmitForm = async () => {
    if (!validateForm()) return

    setLoading(true)
    setError(null)
    try {
      if (dialogMode === 'edit' && editingModel) {
        await updateModel(editingModel.id, {
          name: formData.name,
          api_base_url: formData.api_base_url,
          model_id: formData.model_id,
          api_key: formData.api_key || undefined,
        })
        setSuccessMessage('模型已更新')
      } else {
        await createModel({
          name: formData.name,
          model_type: formData.model_type,
          api_base_url: formData.api_base_url,
          model_id: formData.model_id,
          api_key: formData.api_key,
        })
        setSuccessMessage('模型已添加')
      }

      setDialogOpen(false)
      await loadModels()

      setTimeout(() => setSuccessMessage(null), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存失败')
    } finally {
      setLoading(false)
    }
  }

  /**
   * 打开删除确认对话框
   */
  const handleOpenDeleteDialog = (model: AIModelListItem) => {
    setModelToDelete(model)
    setDeleteDialogOpen(true)
  }

  /**
   * 确认删除模型
   */
  const handleConfirmDelete = async () => {
    if (!modelToDelete) return

    setLoading(true)
    setError(null)
    try {
      await deleteModel(modelToDelete.id)
      setSuccessMessage('模型已删除')
      setDeleteDialogOpen(false)
      setModelToDelete(null)
      await loadModels()

      setTimeout(() => setSuccessMessage(null), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除失败')
    } finally {
      setLoading(false)
    }
  }

  /**
   * 获取模型类型标签
   */
  const getModelTypeBadge = (type: AIModelType) => {
    const typeInfo = MODEL_TYPES.find((t) => t.value === type)
    const variantMap: Record<AIModelType, 'default' | 'secondary' | 'outline'> = {
      qa: 'default',
      embedding: 'secondary',
      rerank: 'outline',
    }
    return (
      <Badge variant={variantMap[type]}>
        {typeInfo?.label ?? type}
      </Badge>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground">模型管理</h1>
        <p className="text-muted-foreground mt-2">
          配置问答、嵌入、重排模型；保存前可用「测试连接」做最小请求校验
        </p>
      </div>

      {successMessage && (
        <div className="p-4 rounded-lg bg-green-500/10 border border-green-500/20">
          <p className="text-sm text-green-600 dark:text-green-400">{successMessage}</p>
        </div>
      )}

      {error && (
        <div className="p-4 rounded-lg bg-destructive/10 border border-destructive/20">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}

      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center justify-between">
            <div className="text-sm text-muted-foreground">
              共 {models.length} 个模型配置
            </div>
            <Button onClick={handleOpenCreateDialog} disabled={loading}>
              添加模型
            </Button>
          </div>
        </CardContent>
      </Card>

      {models.length === 0 && !loading ? (
        <Card>
          <CardContent className="py-12">
            <div className="text-center">
              <p className="text-muted-foreground mb-4">
                暂无模型配置，请点击「添加模型」创建
              </p>
              <Button onClick={handleOpenCreateDialog}>
                添加第一个模型
              </Button>
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {models.map((model) => (
            <Card key={model.id} className="flex flex-col">
              <CardHeader>
                <div className="flex items-start justify-between gap-2">
                  <CardTitle className="line-clamp-1">{model.name}</CardTitle>
                  {getModelTypeBadge(model.model_type)}
                </div>
                <CardDescription className="font-mono text-xs">
                  {model.model_id}
                </CardDescription>
              </CardHeader>
              <CardContent className="flex-1">
                <div className="space-y-2 text-sm">
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">API 地址：</span>
                    <span className="text-foreground text-xs truncate max-w-[200px]" title={model.api_base_url}>
                      {model.api_base_url}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">API Key：</span>
                    <span className={model.has_api_key ? 'text-green-600 dark:text-green-400' : 'text-muted-foreground'}>
                      {model.has_api_key ? '已配置' : '未配置'}
                    </span>
                  </div>
                </div>
              </CardContent>
              <div className="border-t p-4 flex flex-wrap gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleOpenEditDialog(model)}
                >
                  编辑
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={() => handleOpenDeleteDialog(model)}
                >
                  删除
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}

      {/* 创建/编辑模型对话框 */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-[520px]">
          <DialogHeader>
            <DialogTitle>
              {dialogMode === 'create' ? '添加模型' : '编辑模型'}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label htmlFor="model-name" className="text-sm font-medium text-foreground">
                显示名称 <span className="text-destructive">*</span>
              </label>
              <Input
                id="model-name"
                placeholder="输入模型显示名称"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                disabled={loading}
                className={formErrors.name ? 'border-destructive' : ''}
              />
              {formErrors.name && (
                <p className="text-xs text-destructive">{formErrors.name}</p>
              )}
            </div>

            {dialogMode === 'create' && (
              <div className="space-y-2">
                <label htmlFor="model-type" className="text-sm font-medium text-foreground">
                  模型类型
                </label>
                <Select
                  value={formData.model_type}
                  onValueChange={(value) => setFormData({ ...formData, model_type: value as AIModelType })}
                  disabled={loading}
                >
                  <SelectTrigger id="model-type">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {MODEL_TYPES.map((type) => (
                      <SelectItem key={type.value} value={type.value}>
                        {type.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  {MODEL_TYPES.find((t) => t.value === formData.model_type)?.description}
                </p>
              </div>
            )}

            <div className="space-y-2">
              <label htmlFor="api-url" className="text-sm font-medium text-foreground">
                API 地址 <span className="text-destructive">*</span>
              </label>
              <Input
                id="api-url"
                placeholder="https://api.openai.com/v1/embeddings"
                value={formData.api_base_url}
                onChange={(e) => setFormData({ ...formData, api_base_url: e.target.value })}
                disabled={loading}
                className={formErrors.api_base_url ? 'border-destructive' : ''}
              />
              {formErrors.api_base_url && (
                <p className="text-xs text-destructive">{formErrors.api_base_url}</p>
              )}
              <p className="text-xs text-muted-foreground">
                问答：<code className="text-xs">/v1/chat/completions</code> · 嵌入：<code className="text-xs">/v1/embeddings</code> · 重排：<code className="text-xs">/v1/rerank</code>
              </p>
            </div>

            <div className="space-y-2">
              <label htmlFor="model-id" className="text-sm font-medium text-foreground">
                模型 ID <span className="text-destructive">*</span>
              </label>
              <Input
                id="model-id"
                placeholder="text-embedding-3-small"
                value={formData.model_id}
                onChange={(e) => setFormData({ ...formData, model_id: e.target.value })}
                disabled={loading}
                className={formErrors.model_id ? 'border-destructive' : ''}
              />
              {formErrors.model_id && (
                <p className="text-xs text-destructive">{formErrors.model_id}</p>
              )}
              <p className="text-xs text-muted-foreground">
                请求体中的 model 字段值
              </p>
            </div>

            <div className="space-y-2">
              <label htmlFor="api-key" className="text-sm font-medium text-foreground">
                API Key {dialogMode === 'edit' && <span className="text-muted-foreground">(留空表示不修改)</span>}
                {dialogMode === 'create' && <span className="text-destructive">*</span>}
              </label>
              <Input
                id="api-key"
                type="password"
                autoComplete="off"
                placeholder="sk-..."
                value={formData.api_key}
                onChange={(e) => setFormData({ ...formData, api_key: e.target.value })}
                disabled={loading}
                className={formErrors.api_key ? 'border-destructive' : ''}
              />
              {formErrors.api_key && (
                <p className="text-xs text-destructive">{formErrors.api_key}</p>
              )}
              {dialogMode === 'edit' && editingModel?.has_api_key && (
                <p className="text-xs text-muted-foreground">
                  已保存密钥，测试连接可使用已存密钥
                </p>
              )}
            </div>

            {testResult && (
              <div
                className={`p-3 rounded-lg text-sm ${
                  testResult.ok
                    ? 'bg-green-500/10 border border-green-500/20 text-green-600 dark:text-green-400'
                    : 'bg-destructive/10 border border-destructive/20 text-destructive'
                }`}
              >
                {testResult.text}
              </div>
            )}

            <div className="flex gap-2 pt-2">
              <Button
                variant="outline"
                onClick={handleTestConnection}
                disabled={loading || testLoading || !canTestConnection}
              >
                {testLoading ? '测试中...' : '测试连接'}
              </Button>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDialogOpen(false)}
              disabled={loading}
            >
              取消
            </Button>
            <Button
              onClick={handleSubmitForm}
              disabled={loading || !formData.name.trim() || !formData.api_base_url.trim() || !formData.model_id.trim()}
            >
              {loading ? '保存中...' : '保存'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 删除确认对话框 */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
          </DialogHeader>
          <div className="py-4">
            <p className="text-sm text-muted-foreground">
              确定要删除模型「{modelToDelete?.name}」吗？此操作不可恢复。
            </p>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteDialogOpen(false)}
              disabled={loading}
            >
              取消
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirmDelete}
              disabled={loading}
            >
              {loading ? '删除中...' : '确认删除'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}