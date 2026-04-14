# Zustand Auth Store 创建记录

## 完成内容
- 创建 frontend-opencode/src/stores/authStore.ts
- 使用 zustand + persist 中间件实现 localStorage 持久化
- 存储 key: rag_online_course_auth

## 关键模式
- 使用  模式
-  控制哪些状态持久化
- 接口设计与现有 Context 实现保持一致 (AuthUser: id, username, role)

## 使用示例
```tsx
import { useAuthStore } from '@/stores/authStore'

const { token, user, setAuth, logout } = useAuthStore()
```

## 后续
- Task 8 (认证服务) 将使用此 store
- Task 14 (完整Auth Store) 依赖此基础
- Task 21 (登录页) 将调用 setAuth/logout

