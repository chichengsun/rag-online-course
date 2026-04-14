# Frontend Redesign - React Implementation

## TL;DR

> **Quick Summary**: 完整重写前端项目，在frontend-opencode目录创建基于Vite+React+shadcn/ui的现代化前端，实现教师端10页+学生端4页共15个页面，集成Zustand状态管理、Vitest测试框架，复用后端API和现有业务逻辑。
> 
> **Deliverables**: 
> - frontend-opencode项目（Vite + React 19 + TypeScript）
> - 教师端10个页面（认证、课程管理、知识库、对话、AI模型）
> - 学生端4个页面（课程列表、我的课程、课程目录、学习进度）
> - API服务层（fetch封装 + SSE流式支持）
> - Zustand状态管理（auth store + course store）
> - Vitest测试框架配置
> - shadcn/ui组件库集成 + Tailwind CSS
> 
> **Estimated Effort**: Large（13-17天）
> **Parallel Execution**: YES - 5-7 tasks per wave
> **Critical Path**: 项目初始化 → 基础设施 → 教师端页面 → 学生端页面 → 测试

---

## Context

### Original Request
在frontend-opencode目录下创建全新的React前端项目，采用Vite+React+shadcn/ui技术栈，参考后端API实现所有现有功能，完整实现教师端和学生端页面。

### Interview Summary
**Key Discussions**:
- **框架选择**: Vite + React 18 + TypeScript（现代化构建工具，开发体验好）
- **UI方案**: shadcn/ui + Tailwind CSS（高度可定制，现代化设计）
- **状态管理**: Zustand（轻量简洁，替代Context API）
- **测试策略**: Tests-after（先实现后测试，Vitest + @testing-library/react）
- **功能范围**: 教师端10页 + 学生端4页（完整实现现有功能）

**Research Findings**:
- **后端API**: 45+端点，JWT+Redis Session认证，SSE流式支持
- **现有前端**: React 19 + Vite，原生CSS，无测试框架，Context API状态管理
- **技术栈**: react-router-dom，react-markdown，fetch封装
- **学生端缺口**: 现有frontend仅有教师端，学生端需新建4个页面

### Metis Review
**Identified Gaps** (addressed):
- **学生端功能细节**: 课程列表、我的课程、课程目录、学习进度4个页面
- **代码迁移策略**: 迁移services层和auth逻辑，参考现有关键实现
- **组件定制需求**: shadcn/ui默认主题，必要品牌定制（主题色）
- **认证流程细节**: 自动Token刷新、localStorage持久化
- **测试覆盖要求**: Tests-after，核心流程测试，目标覆盖率60%
- **Guardrails**: 不修改现有frontend目录，不新增后端API，不添加国际化/PWA

---

## Work Objectives

### Core Objective
创建全新的现代化前端项目，完整实现RAG在线课程平台的教师端和学生端功能，采用shadcn/ui组件库提升UI质量，使用Zustand简化状态管理。

### Concrete Deliverables
- `frontend-opencode/` 项目目录
- 项目配置文件（package.json, vite.config.ts, tsconfig.json, tailwind.config.js）
- API服务层（`src/services/*.ts`）
- 状态管理（`src/stores/*.ts`）
- 路由配置（`src/router.tsx`）
- 教师端布局（`src/layouts/TeacherLayout.tsx`）
- 学生端布局（`src/layouts/StudentLayout.tsx`）
- 教师端10个页面（`src/pages/teacher/*.tsx`）
- 学生端4个页面（`src/pages/student/*.tsx`）
- 共享组件（`src/components/*.tsx`）
- 测试文件（`src/**/*.test.ts(x)`）

### Definition of Done
- [ ] `npm run dev` 启动开发服务器，访问 http://localhost:5173 显示登录页
- [ ] `npm run build` 构建成功，无TypeScript错误
- [ ] `npm run test` 测试通过，覆盖率 ≥ 60%
- [ ] 教师端10个页面功能完整，API调用正常
- [ ] 学生端4个页面功能完整，API调用正常
- [ ] SSE流式对话正常工作，支持网络断开重连

### Must Have
- 完整的教师端10个页面（AuthPage, TeacherCoursesPage, TeacherCourseContentPage, TeacherResourcePreviewPage, TeacherKnowledgeHubPage, TeacherKnowledgeResourcesPage, TeacherKnowledgeChunkPage, TeacherKnowledgeChatsPage, TeacherKnowledgeChatDetailPage, TeacherAIModelsPage）
- 完整的学生端4个页面（StudentCoursesPage, StudentMyCoursesPage, StudentCourseCatalogPage, StudentProgressPage）
- API服务层（fetch封装 + SSE流式支持）
- Zustand状态管理（auth store + course store）
- 基础测试覆盖（认证、CRUD、SSE）

### Must NOT Have (Guardrails)
- ❌ 修改现有 `frontend/` 目录（代码隔离）
- ❌ 修改后端代码（API边界）
- ❌ 添加国际化（i18n）
- ❌ 添加暗黑模式
- ❌ 添加PWA离线支持
- ❌ 过度定制shadcn/ui组件（使用默认主题）
- ❌ 过度设计测试（Tests-after，核心流程测试即可）
- ❌ 学生端额外功能（学习笔记、讨论区等）

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** - ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: NO（新建项目）
- **Automated tests**: Tests-after
- **Framework**: Vitest + @testing-library/react
- **Coverage target**: 60%（核心流程测试）

### QA Policy
Every task MUST include agent-executed QA scenarios.

- **Frontend/UI**: Use Playwright (playwright skill) - Navigate, interact, assert DOM, screenshot
- **API/Backend**: Use Bash (curl) - Send requests, assert status + response fields
- **Library/Module**: Use Bash (node/vite) - Import, call functions, validate output

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately - foundation + scaffolding):
├── Task 1: 项目初始化 + 配置 [quick]
├── Task 2: shadcn/ui集成 + Tailwind配置 [quick]
├── Task 3: TypeScript类型定义 [quick]
├── Task 4: Zustand store设置 [quick]
├── Task 5: API服务层基础（request封装） [quick]
├── Task 6: 路由配置 + ProtectedRoute [quick]
└── Task 7: Vitest测试配置 [quick]

Wave 2 (After Wave 1 - core services):
├── Task 8: 认证服务层（auth service） [quick]
├── Task 9: 课程服务层（course service） [quick]
├── Task 10: 资源服务层（resource service） [quick]
├── Task 11: 知识库服务层（knowledge service） [quick]
├── Task 12: 对话服务层（chat service + SSE） [deep]
├── Task 13: AI模型服务层（aiModels service） [quick]
└── Task 14: Auth Store完整实现 [quick]

Wave 3 (After Wave 2 - layouts + shared components):
├── Task 15: 共享UI组件（Button, Card, Input等） [visual-engineering]
├── Task 16: 教师端布局（TeacherLayout + 侧边栏） [visual-engineering]
├── Task 17: 学生端布局（StudentLayout + 导航） [visual-engineering]
├── Task 18: 文件上传组件（Upload组件 + 进度条） [visual-engineering]
├── Task 19: Markdown渲染组件 [quick]
└── Task 20: SSE消息组件（流式显示） [visual-engineering]

Wave 4 (After Wave 3 - teacher auth + courses):
├── Task 21: 登录/注册页面（AuthPage） [visual-engineering]
├── Task 22: 课程列表页（TeacherCoursesPage） [visual-engineering]
├── Task 23: 课程创建/编辑 [visual-engineering]
├── Task 24: 章节管理（Chapter CRUD） [visual-engineering]
├── Task 25: 节管理（Section CRUD） [visual-engineering]
└── Task 26: 资源管理（Resource CRUD） [unspecified-high]

Wave 5 (After Wave 4 - teacher knowledge + chat):
├── Task 27: 资源预览页（PDF/Video/Audio） [visual-engineering]
├── Task 28: 知识库入口页（TeacherKnowledgeHubPage） [visual-engineering]
├── Task 29: 知识库资源列表（TeacherKnowledgeResourcesPage） [visual-engineering]
├── Task 30: 知识库分块管理（TeacherKnowledgeChunkPage） [unspecified-high]
├── Task 31: 对话列表页（TeacherKnowledgeChatsPage） [visual-engineering]
└── Task 32: 对话详情页（SSE流式） [deep]

Wave 6 (After Wave 5 - teacher AI models + student pages):
├── Task 33: AI模型管理页（TeacherAIModelsPage） [visual-engineering]
├── Task 34: 课程列表页（StudentCoursesPage） [visual-engineering]
├── Task 35: 我的课程页（StudentMyCoursesPage） [visual-engineering]
├── Task 36: 课程目录页（StudentCourseCatalogPage） [visual-engineering]
└── Task 37: 学习进度页（StudentProgressPage） [visual-engineering]

Wave FINAL (After ALL tasks — 4 parallel reviews, then user okay):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)
-> Present results -> Get explicit user okay

Critical Path: Task 1 → Task 5 → Task 12 → Task 32 → Wave FINAL
Parallel Speedup: ~60% faster than sequential
Max Concurrent: 7 (Wave 1 & 2)
```

---

## TODOs

- [x] 1. **项目初始化 + 配置**

  **What to do**:
  - 在项目根目录创建 `frontend-opencode/` 目录
  - 使用 `npm create vite@latest frontend-opencode -- --template react-ts` 初始化项目
  - 安装核心依赖：`npm install react-router-dom zustand`
  - 配置 `vite.config.ts`（API代理、别名）
  - 配置 `tsconfig.json`（严格模式、路径别名）
  - 创建 `.env.local` 环境变量文件（API_BASE_URL）

  **Must NOT do**:
  - 不要修改现有 `frontend/` 目录
  - 不要使用 CRA (Create React App)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 标准的项目初始化，配置文件模板化
  - **Skills**: []
  - **Skills Evaluated but Omitted**: `vercel-react-best-practices` (初始化阶段不需要)

  **Parallelization**:
  - **Can Run In Parallel**: NO (依赖项目)
  - **Parallel Group**: Wave 1 (启动)
  - **Blocks**: Task 2-7
  - **Blocked By**: None (可立即开始)

  **References**:
  - Pattern: `frontend/vite.config.ts` - Vite配置模式（代理、别名）
  - Pattern: `frontend/tsconfig.json` - TypeScript配置模式
  - Pattern: `frontend/.env` - 环境变量文件模式

  **Acceptance Criteria**:
  - [ ] 目录 `frontend-opencode/` 存在
  - [ ] `npm run dev` 成功启动开发服务器
  - [ ] 访问 http://localhost:5173 显示Vite默认页面
  - [ ] `package.json` 包含 react-router-dom、zustand 依赖
  - [ ] `vite.config.ts` 配置了API代理（target: http://localhost:8080）
  - [ ] `tsconfig.json` 配置了路径别名（@/src）

  **QA Scenarios**:
  ```bash
  Scenario: Dev server starts successfully
    Tool: Bash
    Preconditions: Node.js 18+ installed
    Steps:
      1. cd frontend-opencode
      2. npm install
      3. npm run dev
    Expected Result: Server starts on port 5173, no errors
    Failure Indicators: Port conflict, dependency install fails
    Evidence: .sisyphus/evidence/task-1-dev-start.log

  Scenario: Technology stack validation
    Tool: Bash
    Preconditions: Project initialized
    Steps:
      1. cat frontend-opencode/package.json | grep -E "(react|vite|zustand|react-router)"
    Expected Result: All dependencies present with correct versions
    Failure Indicators: Missing dependencies
    Evidence: .sisyphus/evidence/task-1-deps.log
  ```

  **Commit**: YES (groups with Wave 1)
  - Message: `chore(init): initialize Vite + React project`
  - Files: `package.json, vite.config.ts, tsconfig.json, .env.local`
  - Pre-commit: `npm run dev` (验证项目启动)

- [x] 2. **shadcn/ui集成 + Tailwind配置**

  **What to do**:
  - 安装Tailwind CSS：`npm install -D tailwindcss postcss autoprefixer`
  - 初始化Tailwind：`npx tailwindcss init -p`
  - 安装shadcn/ui CLI：`npx shadcn@latest init`
  - 配置 `tailwind.config.js`（content paths, theme colors）
  - 配置 `src/index.css`（Tailwind directives）
  - 安装基础组件：`npx shadcn@latest add button card input`

  **Must NOT do**:
  - 不要过度定制主题（使用默认主题）
  - 不要安装所有shadcn组件（仅安装需要的）

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: UI/样式配置需要设计sense
  - **Skills**: [`frontend-design`]
    - `frontend-design`: shadcn/ui配置和Tailwind集成

  **Parallelization**:
  - **Can Run In Parallel**: NO (依赖Task 1完成)
  - **Parallel Group**: Wave 1
  - **Blocks**: Task 15
  - **Blocked By**: Task 1

  **References**:
  - External: `https://ui.shadcn.com/docs/installation/vite` - shadcn/ui Vite安装指南
  - External: `https://tailwindcss.com/docs/installation` - Tailwind CSS安装指南

  **Acceptance Criteria**:
  - [ ] `tailwind.config.js` 配置正确（content包含src目录）
  - [ ] `src/index.css` 包含Tailwind directives
  - [ ] `npx shadcn@latest add button` 成功安装Button组件
  - [ ] Button组件可在页面中使用：<Button>Test</Button>
  - [ ] Tailwind样式正常工作（class="text-blue-500"生效）

  **QA Scenarios**:
  ```bash
  Scenario: Tailwind CSS works
    Tool: Bash
    Preconditions: Tailwind installed and configured
    Steps:
      1. Add class="text-blue-500" to a component
      2. Render component in browser
      3. Inspect element to verify text color
    Expected Result: Text is blue (#3b82f6)
    Failure Indicators: No color change, style not applied
    Evidence: .sisyphus/evidence/task-2-tailwind.png

  Scenario: shadcn/ui Button renders
    Tool: Bash
    Preconditions: shadcn/ui installed
    Steps:
      1. Import Button from "@/components/ui/button"
      2. Render <Button variant="default">Test</Button>
      3. Verify button appears with correct styles
    Expected Result: Button renders with shadcn default styles
    Failure Indicators: Import errors, rendering fails
    Evidence: .sisyphus/evidence/task-2-shadcn-button.png
  ```

  **Commit**: YES (groups with Wave 1)
  - Message: `feat(ui): integrate shadcn/ui and Tailwind CSS`
  - Files: `tailwind.config.js, postcss.config.js, src/index.css, src/components/ui/`
  - Pre-commit: 检查Button组件可渲染

- [x] 3. **TypeScript类型定义**

  **What to do**:
  - 创建 `src/types/` 目录
  - 定义API响应类型：`ApiResponse<T>`, `PaginatedResponse<T>`, `ApiError`
  - 定义业务实体类型：`User`, `Course`, `Chapter`, `Section`, `Resource`, `KnowledgeChunk`, `ChatSession`, `ChatMessage`, `AIModel`
  - 定义DTO类型：CreateCourseReq, UpdateCourseReq, LoginReq, RegisterReq等
  - 导出所有类型：`src/types/index.ts`

  **Must NOT do**:
  - 不要使用 `any` 类型
  - 不要定义后端不存在的字段

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 类型定义是模板化工作
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (与Task 1-2独立)
  - **Parallel Group**: Wave 1
  - **Blocks**: Task 8-13 (服务层)
  - **Blocked By**: Task 1 (项目初始化)

  **References**:
  - External: `https://www.typescriptlang.org/docs/handbook/2/types-from-types.html` - TypeScript类型定义指南
  - Pattern: `frontend/src/services/api.ts` - API封装类型参考

  **Acceptance Criteria**:
  - [ ] `src/types/` 目录存在
  - [ ] 所有核心实体类型定义完整（User, Course, Chapter, Section, Resource, KnowledgeChunk, ChatSession, ChatMessage, AIModel）
  - [ ] API响应类型定义（ApiResponse<T>, PaginatedResponse<T>, ApiError）
  - [ ] `npm run build` 无TypeScript错误
  - [ ] 类型可正确导入使用：`import { User } from '@/types'`

  **QA Scenarios**:
  ```bash
  Scenario: TypeScript compilation succeeds
    Tool: Bash
    Preconditions: Types defined
    Steps:
      1. npm run build
    Expected Result: Build succeeds with no TypeScript errors
    Failure Indicators: Type errors in compilation
    Evidence: .sisyphus/evidence/task-3-build.log

  Scenario: Types are usable
    Tool: Bash
    Preconditions: Types defined
    Steps:
      1. Create test.ts file importing all types
      2. Use each type in a mock function
      3. Run tsc --noEmit
    Expected Result: No type errors
    Failure Indicators: Import errors, missing fields
    Evidence: .sisyphus/evidence/task-3-types-usable.log
  ```

  **Commit**: YES (groups with Wave 1)
  - Message: `feat(types): add TypeScript type definitions`
  - Files: `src/types/`
  - Pre-commit: `npm run build` (验证类型编译通过)

- [x] 4. **Zustand store设置**

  **What to do**:
  - 创建 `src/stores/` 目录
  - 创建基础store结构：`useAuthStore`, `useCourseStore`, `useChatStore`
  - 定义auth store：`token`, `user`, `setAuth()`, `logout()`
  - 定义course store：`currentCourse`, `setCurrentCourse()`
  - 定义chat store：`currentSession`, `messages`, `addMessage()`
  - 配置localStorage持久化（token、user）

  **Must NOT do**:
  - 不要在每个组件中创建store（仅全局状态使用Zustand）
  - 不要使用Redux模式（保持Zustand简洁）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Zustand配置简单，模式清晰
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (与Task 3独立)
  - **Parallel Group**: Wave 1
  - **Blocks**: Task 8 (认证服务层), Task 14 (完整Auth Store)
  - **Blocked By**: Task 1

  **References**:
  - External: `https://zustand-demo.pmnd.rs/` - Zustand官方demo
  - External: `https://docs.pmnd.rs/zustand/guides/persisting-data` - Zustand持久化指南
  - Pattern: `frontend/src/store/auth.tsx` - 现有认证状态管理模式

  **Acceptance Criteria**:
  - [ ] `src/stores/` 目录存在
  - [ ] `useAuthStore` 返回 token, user, setAuth, logout
  - [ ] `useCourseStore` 返回 currentCourse, setCurrentCourse
  - [ ] `useChatStore` 返回 currentSession, messages, addMessage
  - [ ] Token持久化到localStorage（key: rag_online_course_access_token）
  - [ ] Store可在组件中使用：`const { token } = useAuthStore()`

  **QA Scenarios**:
  ```bash
  Scenario: Auth store persists token
    Tool: Bash
    Preconditions: Stores created
    Steps:
      1. Render component using useAuthStore
      2. Call setAuth({ token: 'test-token', user: {...} })
      3. Verify localStorage contains 'rag_online_course_access_token'
      4. Refresh page
      5. Verify token still accessible via useAuthStore
    Expected Result: Token survives page refresh
    Failure Indicators: Token lost on refresh
    Evidence: .sisyphus/evidence/task-4-auth-persist.png

  Scenario: Store usage in component
    Tool: Bash
    Preconditions: Stores created
    Steps:
      1. Create test component: const { token } = useAuthStore()
      2. Run component in browser
      3. Verify store values accessible
    Expected Result: Store hooks work correctly
    Failure Indicators: Hook errors, undefined values
    Evidence: .sisyphus/evidence/task-4-store-usage.png
  ```

  **Commit**: YES (groups with Wave 1)
  - Message: `feat(stores): setup Zustand stores`
  - Files: `src/stores/`
  - Pre-commit: 检查store可在组件中导入使用

- [x] 5. **API服务层基础（request封装）**

  **What to do**:
  - 创建 `src/services/api.ts`
  - 实现 `request<T>(path, options)` 通用请求方法
  - 自动附加 `Authorization: Bearer {token}` 头
  - 实现API错误处理类 `ApiError`
  - 实现JSON序列化/反序列化
  - 实现响应拦截器（错误统一处理）
  - 配置Vite proxy：`/api/v1` → `http://localhost:8080`

  **Must NOT do**:
  - 不要在service层引入UI逻辑（Toast等）
  - 不要硬编码API基础URL（使用环境变量）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: HTTP封装模式固定，可参考现有实现
  - **Skills**: []
  - **Skills Evaluated but Omitted**: `vercel-react-best-practices` (这是基础设施层)

  **Parallelization**:
  - **Can Run In Parallel**: YES (与Task 3-4独立)
  - **Parallel Group**: Wave 1
  - **Blocks**: Task 8-13 (各服务层)
  - **Blocked By**: Task 1

  **References**:
  - Pattern: `frontend/src/services/api.ts` - 现有API封装实现
  - Pattern: `frontend/vite.config.ts` - 代理配置
  - External: `https://vitejs.dev/config/server-options.html#server-proxy` - Vite代理配置

  **Acceptance Criteria**:
  - [ ] `src/services/api.ts` 存在
  - [ ] `request<T>(path, options)` 泛型方法可调用
  - [ ] `ApiError` 类可抛出，包含 `code`, `message`, `status`
  - [ ] 自动附加 Authorization 头（当token存在时）
  - [ ] Vite proxy配置正确
  - [ ] 错误响应被正确处理

  **QA Scenarios**:
  ```bash
  Scenario: Request adds auth header
    Tool: Bash
    Preconditions: API service created
    Steps:
      1. Set token via useAuthStore.getState().setAuth({ token: 'test-token' })
      2. Call request('/api/v1/me', { method: 'GET' })
      3. Verify request includes Authorization header
    Expected Result: Header present: Authorization: Bearer test-token
    Failure Indicators: Header missing
    Evidence: .sisyphus/evidence/task-5-auth-header.log

  Scenario: API error handling
    Tool: Bash
    Preconditions: API service created
    Steps:
      1. Call request('/api/v1/nonexistent', { method: 'GET' })
      2. Catch ApiError
      3. Verify error contains code, message, status
    Expected Result: ApiError thrown with correct fields
    Failure Indicators: Generic error instead of ApiError
    Evidence: .sisyphus/evidence/task-5-error-handling.log
  ```

  **Commit**: YES (groups with Wave 1)
  - Message: `feat(services): add base API request wrapper`
  - Files: `src/services/api.ts, vite.config.ts`
  - Pre-commit: 验证mock请求可发起

- [x] 6. **路由配置 + ProtectedRoute**

  **What to do**:
  - 安装 `react-router-dom`（已在Task 1中安装）
  - 创建 `src/router.tsx` 路由配置
  - 定义路由结构：
    - `/auth` → AuthPage
    - `/teacher/*` → TeacherLayout (Protected)
    - `/student/*` → StudentLayout (Protected)
  - 实现 `ProtectedRoute` 组件（检查token，未登录重定向到/auth）
  - 在 `App.tsx` 中使用 `<RouterProvider>`
  - 配置路由守卫：`useAuthStore()` 检查token

  **Must NOT do**:
  - 不要在路由中硬编码页面组件名称（使用lazy loading）
  - 不要在ProtectedRoute中引入UI逻辑

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 路由配置模式固定
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (与Task 5独立)
  - **Parallel Group**: Wave 1
  - **Blocks**: Task 21-37 (所有页面)
  - **Blocked By**: Task 1, Task 4

  **References**:
  - Pattern: `frontend/src/App.tsx` - 现有路由配置
  - External: `https://reactrouter.com/en/main/start/overview` - React Router v7文档

  **Acceptance Criteria**:
  - [ ] `src/router.tsx` 存在
  - [ ] `ProtectedRoute` 组件可检查token并重定向
  - [ ] 路由结构正确（/auth, /teacher/*, /student/*）
  - [ ] 未登录访问 `/teacher/*` 自动重定向到 `/auth`
  - [ ] 已登录访问 `/auth` 自动重定向到 `/teacher/courses`

  **QA Scenarios**:
  ```bash
  Scenario: ProtectedRoute redirects unauthenticated users
    Tool: Bash
    Preconditions: Router configured
    Steps:
      1. Clear localStorage (no token)
      2. Navigate to /teacher/courses
      3. Verify redirect to /auth
    Expected Result: URL is /auth after redirect
    Failure Indicators: Shows teacher page without auth
    Evidence: .sisyphus/evidence/task-6-protected-redirect.png

  Scenario: ProtectedRoute allows authenticated users
    Tool: Bash
    Preconditions: Router configured
    Steps:
      1. Set token via useAuthStore
      2. Navigate to /teacher/courses
      3. Verify page loads (not redirected)
    Expected Result: Teacher courses page renders
    Failure Indicators: Unexpected redirect to /auth
    Evidence: .sisyphus/evidence/task-6-authenticated-access.png
  ```

  **Commit**: YES (groups with Wave 1)
  - Message: `feat(router): add routing configuration and ProtectedRoute`
  - Files: `src/router.tsx, src/components/ProtectedRoute.tsx, src/App.tsx`
  - Pre-commit: `npm run dev` (验证路由可访问)

- [x] 7. **Vitest测试配置**

  **What to do**:
  - 安装测试依赖：`npm install -D vitest @testing-library/react @testing-library/jest-dom jsdom @vitest/ui`
  - 创建 `vitest.config.ts` 配置文件
  - 创建 `src/test/setup.ts` 测试设置文件（配置jest-dom匹配器）
  - 在 `package.json` 添加测试脚本：
    - `"test": "vitest"`
    - `"test:ui": "vitest --ui"`
    - `"test:coverage": "vitest run --coverage"`
  - 创建示例测试：`src/App.test.tsx`

  **Must NOT do**:
  - 不要追求100%覆盖率（Tests-after，核心流程60%即可）
  - 不要配置复杂的测试环境（保持简洁）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 测试配置模板化
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (与所有Task独立)
  - **Parallel Group**: Wave 1
  - **Blocks**: 无（测试在最后阶段）
  - **Blocked By**: Task 1

  **References**:
  - External: `https://vitest.dev/config/` - Vitest配置文档
  - External: `https://testing-library.com/docs/react-testing-library/setup/` - Testing Library设置

  **Acceptance Criteria**:
  - [ ] `vitest.config.ts` 存在
  - [ ] `src/test/setup.ts` 存在
  - [ ] `npm run test` 成功运行（即使无测试用例）
  - [ ] `@testing-library/jest-dom` 匹配器可用
  - [ ] 测试脚本在 package.json 中定义

  **QA Scenarios**:
  ```bash
  Scenario: Vitest runs successfully
    Tool: Bash
    Preconditions: Vitest configured
    Steps:
      1. npm run test
    Expected Result: Vitest runs, shows "No tests found" (acceptable)
    Failure Indicators: Configuration errors, import errors
    Evidence: .sisyphus/evidence/task-7-vitest-run.log

  Scenario: Test setup works
    Tool: Bash
    Preconditions: Setup file created
    Steps:
      1. Create test file with jest-dom matcher: expect(element).toBeInTheDocument()
      2. Run npm run test
    Expected Result: Test passes with no errors
    Failure Indicators: Matcher not recognized
    Evidence: .sisyphus/evidence/task-7-test-setup.log
  ```

  **Commit**: YES (groups with Wave 1)
  - Message: `test(setup): configure Vitest and Testing Library`
  - Files: `vitest.config.ts, src/test/setup.ts, package.json`
  - Pre-commit: `npm run test` (验证配置正确)

**Commit**: YES (groups with Wave 1)
  - Message: `test(setup): configure Vitest and Testing Library`
  - Files: `vitest.config.ts, src/test/setup.ts, package.json`
  - Pre-commit: `npm run test` (验证配置正确)

- [x] 8. **认证服务层（auth service）**

  **What to do**:
  - 创建 `src/services/auth.ts`
  - 实现 `register(data: RegisterReq)` 注册方法
  - 实现 `login(data: LoginReq)` 登录方法
  - 实现 `refreshToken(refresh_token: string)` 刷新Token方法
  - 实现 `logout()` 登出方法（清理localStorage）
  - 定义LoginReq, RegisterReq, LoginResp类型（参考backend API）

  **Must NOT do**:
  - 不要在service层引入状态管理（Zustand）
  - 不要存储敏感信息（密码明文等）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: API封装模式固定
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (与Task 9-13并行)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 14 (完整Auth Store), Task 21 (登录页)
  - **Blocked By**: Task 3 (类型定义), Task 5 (API基础)

  **References**:
  - Pattern: `frontend/src/services/auth.ts` - 现有认证服务实现
  - Backend: `POST /api/v1/auth/register`, `/login`, `/refresh` - API端点

  **Acceptance Criteria**:
  - [ ] `src/services/auth.ts` 存在
  - [ ] `register()` 调用 `/api/v1/auth/register`
  - [ ] `login()` 调用 `/api/v1/auth/login`，返回 `{ access_token, refresh_token, user }`
  - [ ] `refreshToken()` 调用 `/api/v1/auth/refresh`
  - [ ] 所有方法返回Promise，类型正确

  **QA Scenarios**:
  ```bash
  Scenario: Login API call succeeds
    Tool: Bash
    Preconditions: Backend running on localhost:8080
    Steps:
      1. Call login({ account: 'test@example.com', password: 'password' })
      2. Verify response contains access_token and refresh_token
    Expected Result: Response includes tokens and user object
    Failure Indicators: API error, missing tokens
    Evidence: .sisyphus/evidence/task-8-login.log
  ```

  **Commit**: YES (groups with Wave 2)
  - Message: `feat(services): add auth service layer`
  - Files: `src/services/auth.ts`
  - Pre-commit: 无

- [x] 9. **课程服务层（course service）**

  **What to do**:
  - 创建 `src/services/course.ts`
  - 实现 `getCourses(params)` 课程列表（分页、筛选）
  - 实现 `createCourse(data)` 创建课程
  - 实现 `updateCourse(courseId, data)` 更新课程
  - 实现 `deleteCourse(courseId)` 删除课程
  - 实现 `getChapters(courseId)` 获取章节列表
  - 实现 `createChapter(courseId, data)` 创建章节
  - 实现 `getSections(courseId, chapterId)` 获取节列表
  - 实现 `createSection(courseId, chapterId, data)` 创建节
  - 定义Course, Chapter, Section类型

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: CRUD操作模式固定
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (与Task 8, 10-13并行)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 22-26 (课程管理页面)
  - **Blocked By**: Task 3, Task 5

  **References**:
  - Pattern: `frontend/src/services/course.ts` - 现有课程服务
  - Backend: `/api/v1/teacher/courses`, `/chapters`, `/sections` - API端点

  **Acceptance Criteria**:
  - [ ] 所有CRUD方法实现完整
  - [ ] 返回类型正确（Promise<Course[]>, Promise<Chapter[]>等）
  - [ ] 错误处理正确（抛出ApiError）

  **QA Scenarios**:
  ```bash
  Scenario: Course CRUD works
    Tool: Bash
    Preconditions: Backend running, auth token set
    Steps:
      1. Create course: createCourse({ title: 'Test' })
      2. Get courses: getCourses()
      3. Update course: updateCourse(id, { title: 'Updated' })
      4. Delete course: deleteCourse(id)
    Expected Result: All operations succeed
    Failure Indicators: Any operation fails
    Evidence: .sisyphus/evidence/task-9-course-crud.log
  ```

- [x] 10. **资源服务层（resource service）**

  **What to do**:
  - 创建 `src/services/resource.ts`
  - 实现 `initUpload(data)` 初始化上传（获取预签名URL）
  - 实现 `confirmResource(sectionId, data)` 确认资源入库
  - 实现 `getResource(resourceId)` 获取资源详情
  - 实现 `getPreviewUrl(resourceId)` 获取预览URL
  - 实现 `parseResource(resourceId)` 解析资源
  - 实现 `summarizeResource(resourceId)` 生成AI摘要
  - 实现 `deleteResource(resourceId)` 删除资源
  - 定义InitUploadReq, InitUploadResp, ConfirmResourceReq类型

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 文件上传逻辑较复杂，需要处理预签名URL
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (与Task 8-9, 11-13并行)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 26 (资源管理), Task 27 (资源预览)
  - **Blocked By**: Task 3, Task 5

  **References**:
  - Pattern: `frontend/src/services/course.ts` - 现有资源服务（集成在course中）
  - Backend: `/api/v1/teacher/resources/*` - API端点

  **Acceptance Criteria**:
  - [ ] 所有资源操作方法实现
  - [ ] `initUpload` 返回预签名URL和object_key
  - [ ] `confirmResource` 支持文件元信息确认
  - [ ] 预签名URL上传流程正确

- [x] 11. **知识库服务层（knowledge service）**

  **What to do**:
  - 创建 `src/services/knowledge.ts`
  - 实现 `getKnowledgeResources(courseId)` 获取知识库资源列表
  - 实现 `chunkPreview(resourceId, data)` 分块预览
  - 实现 `saveChunks(resourceId, data)` 保存分块
  - 实现 `confirmChunks(resourceId)` 确认分块
  - 实现 `getChunks(resourceId)` 获取分块列表
  - 实现 `embedResource(resourceId, data)` 嵌入向量
  - 实现 `deleteChunk(resourceId, chunkId)` 删除分块
  - 实现 `updateChunk(resourceId, chunkId, data)` 更新分块
  - 定义ChunkPreviewReq, SaveChunksReq, EmbedResourceReq类型

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 知识库操作逻辑复杂，涉及分块和嵌入
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (与Task 8-10, 12-13并行)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 28-30 (知识库页面)
  - **Blocked By**: Task 3, Task 5

  **References**:
  - Pattern: `frontend/src/services/knowledge.ts` - 现有知识库服务
  - Backend: `/api/v1/teacher/resources/:id/knowledge/*` - API端点

  **Acceptance Criteria**:
  - [ ] 所有知识库操作方法实现
  - [ ] 分块预览、保存、确认流程完整
  - [ ] 嵌入向量方法正确调用 /embed 端点

- [x] 12. **对话服务层（chat service + SSE）**

  **What to do**:
  - 创建 `src/services/chat.ts`
  - 实现 `createSession(courseId, data)` 创建对话会话
  - 实现 `getSessions()` 获取会话列表
  - 实现 `updateSession(sessionId, data)` 更新会话标题
  - 实现 `deleteSession(sessionId)` 删除会话
  - 实现 `getMessages(sessionId)` 获取消息历史
  - 实现 `askInSession(sessionId, data)` 问答（非流式）
  - **实现 `askInSessionStream(sessionId, data, onToken, onReferences, onDone, onError)` 流式问答**
  - 定义CreateSessionReq, AskInSessionReq, AskInSessionResp类型

  **Must NOT do**:
  - 不要使用WebSocket（后端使用SSE）
  - 不要阻塞UI（流式输出需要异步处理）

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: SSE流式实现复杂，需要处理网络断开、重连
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (依赖Task 5的API基础，需要特殊处理流式)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 31-32 (对话页面)
  - **Blocked By**: Task 3, Task 5

  **References**:
  - Pattern: `frontend/src/services/knowledgeChat.ts` - 现有对话服务（SSE实现）
  - Backend: `/api/v1/teacher/knowledge/chats/sessions/:id/ask/stream` - SSE端点
  - External: `https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events` - SSE文档

  **Acceptance Criteria**:
  - [ ] 所有对话操作方法实现
  - [ ] `askInSessionStream` 使用fetch + ReadableStream实现SSE
  - [ ] 支持 `onToken(token)` 回调接收流式token
  - [ ] 支持 `onReferences(references)` 回调接收引用
  - [ ] 支持 `onDone(sessionId, messageId)` 回调标记完成
  - [ ] 支持 `onError(error)` 回调处理错误

  **QA Scenarios**:
  ```bash
  Scenario: SSE stream receives tokens
    Tool: Bash
    Preconditions: Backend running, auth token set
    Steps:
      1. Create session
      2. Call askInSessionStream with "Hello" question
      3. Verify onToken called multiple times with tokens
      4. Verify onReferences called with chunks
      5. Verify onDone called
    Expected Result: Stream works, tokens received incrementally
    Failure Indicators: No tokens, stream fails immediately
    Evidence: .sisyphus/evidence/task-12-sse-stream.log
  ```

- [x] 13. **AI模型服务层（aiModels service）**

  **What to do**:
  - 创建 `src/services/aiModels.ts`
  - 实现 `getModels()` 获取模型列表
  - 实现 `createModel(data)` 创建模型
  - 实现 `updateModel(modelId, data)` 更新模型
  - 实现 `deleteModel(modelId)` 删除模型
  - 实现 `testConnection(data)` 测试连接
  - 定义CreateAIModelReq, UpdateAIModelReq, TestConnectionReq类型

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: CRUD操作模式固定
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (与Task 8-12并行)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 33 (AI模型管理页)
  - **Blocked By**: Task 3, Task 5

  **References**:
  - Pattern: `frontend/src/services/aiModels.ts` - 现有AI模型服务
  - Backend: `/api/v1/teacher/ai-models` - API端点

- [x] 14. **Auth Store完整实现**

  **What to do**:
  - 扩展 `src/stores/authStore.ts`（Task 4已创建基础）
  - 添加 `isAuthenticated()` 方法（检查token是否存在）
  - 添加 `getUser()` 方法（从localStorage恢复user）
  - 实现 `login(email, password)` 方法（调用authService.login + 更新store）
  - 实现 `logout()` 方法（调用authService.logout + 清除store + localStorage）
  - 添加 `loading` 状态
  - 添加 `error` 状态
  - 实现Token自动刷新逻辑（可选，基于refresh_token）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Zustand store扩展简单
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (依赖Task 8认证服务层)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 21 (登录页)
  - **Blocked By**: Task 4, Task 8

  **References**:
  - Pattern: `frontend/src/store/auth.tsx` - 现有Auth Store实现

  **Acceptance Criteria**:
  - [ ] `isAuthenticated()` 正确检查token
  - [ ] `login(email, password)` 调用authService并更新store
  - [ ] `logout()` 清除store和localStorage
  - [ ] `loading` 和 `error` 状态可追踪登录过程
  - [ ] Token从localStorage正确恢复

**Acceptance Criteria**:
  - [ ] `isAuthenticated()` 正确检查token
  - [ ] `login(email, password)` 调用authService并更新store
  - [ ] `logout()` 清除store和localStorage
  - [ ] `loading` 和 `error` 状态可追踪登录过程
  - [ ] Token从localStorage正确恢复

  **QA Scenarios**:
  ```bash
  Scenario: Login updates store correctly
    Tool: Bash
    Preconditions: Auth store created
    Steps:
      1. Call useAuthStore.getState().login('test@example.com', 'password')
      2. Verify token set in store
      3. Verify user object present
      4. Verify localStorage contains token
    Expected Result: All state updated correctly
    Failure Indicators: Token not set, user missing
    Evidence: .sisyphus/evidence/task-14-auth-store-login.png
  ```

- [x] 15. **共享UI组件（Button, Card, Input等）**

  **What to do**:
  - 安装shadcn/ui基础组件：
    - `npx shadcn@latest add button`
    - `npx shadcn@latest add card`
    - `npx shadcn@latest add input`
    - `npx shadcn@latest add textarea`
    - `npx shadcn@latest add select`
    - `npx shadcn@latest add dialog`
    - `npx shadcn@latest add toast`
  - 创建 `src/components/ui/` 目录（shadcn自动创建）
  - 验证所有组件可正常导入使用

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: UI组件集成需要设计sense
  - **Skills**: [`frontend-design`]
    - `frontend-design`: shadcn/ui组件安装和配置

  **Parallelization**:
  - **Can Run In Parallel**: NO (依赖Task 2 shadcn/ui初始化)
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 16-20, Task 21-37 (所有页面)
  - **Blocked By**: Task 2

  **References**:
  - External: `https://ui.shadcn.com/docs/components` - shadcn组件文档

  **Acceptance Criteria**:
  - [ ] 所有组件安装成功
  - [ ] 组件可导入：`import { Button } from '@/components/ui/button'`
  - [ ] 组件可渲染，样式正确

- [x] 16. **教师端布局（TeacherLayout + 侧边栏）**

  **What to do**:
  - 创建 `src/layouts/TeacherLayout.tsx`
  - 实现侧边栏导航：课程管理、知识库、对话、AI模型
  - 实现顶部栏：用户信息、登出按钮
  - 使用 `<Outlet />` 渲染子路由
  - 实现响应式布局（桌面端侧边栏展开，移动端折叠）
  - 样式使用Tailwind CSS + shadcn/ui组件

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: 布局设计需要UI/UX sense
  - **Skills**: [`frontend-design`, `vercel-react-best-practices`]

  **Parallelization**:
  - **Can Run In Parallel**: YES (与Task 17-20并行)
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 22-33 (教师端页面)
  - **Blocked By**: Task 2, Task 6, Task 15

  **References**:
  - Pattern: `frontend/src/layouts/TeacherLayout.tsx` - 现有教师端布局
  - External: `https://ui.shadcn.com/examples/dashboard` - Dashboard布局示例

  **Acceptance Criteria**:
  - [ ] 侧边栏包含所有导航链接（/teacher/courses, /teacher/knowledge, /teacher/chats, /teacher/ai-models）
  - [ ] 顶部栏显示用户信息和登出按钮
  - [ ] `<Outlet />` 正确渲染子路由
  - [ ] 响应式布局正常工作

  **QA Scenarios**:
  ```bash
  Scenario: Sidebar navigation works
    Tool: Bash
    Preconditions: Layout created
    Steps:
      1. Navigate to /teacher/courses
      2. Click "知识库" in sidebar
      3. Verify URL changes to /teacher/knowledge
    Expected Result: Navigation works, page updates
    Evidence: .sisyphus/evidence/task-16-sidebar-nav.png
  ```

- [x] 17. **学生端布局（StudentLayout + 导航）**

  **What to do**:
  - 创建 `src/layouts/StudentLayout.tsx`
  - 实现顶部导航：课程列表、我的课程
  - 使用 `<Outlet />` 渲染子路由
  - 实现简洁的学生端UI风格
  - 样式使用Tailwind CSS + shadcn/ui组件

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`frontend-design`]

  **Parallelization**:
  - **Can Run In Parallel**: YES (与Task 16, 18-20并行)
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 34-37 (学生端页面)
  - **Blocked By**: Task 2, Task 6, Task 15

  **Acceptance Criteria**:
  - [ ] 顶部导航包含所有链接（/student/courses, /student/my-courses）
  - [ ] `<Outlet />` 正确渲染子路由
  - [ ] 学生端UI风格简洁明快

- [x] 18. **文件上传组件（Upload组件 + 进度条）**

  **What to do**:
  - 创建 `src/components/Upload.tsx`
  - 实现文件选择（支持点击和拖拽）
  - 实现文件类型校验（.pdf, .doc, .docx, .ppt, .txt, .mp4, .mp3等）
  - 实现上传进度条（显示百分比）
  - 实现上传成功/失败状态展示
  - 集成MinIO预签名URL上传流程：
    1. 调用 `initUpload` 获取预签名URL
    2. PUT文件到预签名URL
    3. 调用 `confirmResource` 确认上传

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: 复杂交互，需要进度条、拖拽等UI
  - **Skills**: [`frontend-design`]

  **Parallelization**:
  - **Can Run In Parallel**: YES (与Task 16-17, 19-20并行)
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 26 (资源管理)
  - **Blocked By**: Task 10, Task 15

  **References**:
  - External: `https://developer.mozilla.org/en-US/docs/Web/API/File_API/Using_files_from_web_applications` - File API文档
  - External: `https://ui.shadcn.com/docs/components/progress` - Progress组件

  **Acceptance Criteria**:
  - [ ] 支持点击和拖拽上传
  - [ ] 文件类型校验正确
  - [ ] 进度条显示上传百分比
  - [ ] 成功/失败状态正确展示
  - [ ] 预签名URL上传流程正确

- [x] 19. **Markdown渲染组件**

  **What to do**:
  - 安装 `react-markdown` 和 `remark-gfm`
  - 创建 `src/components/MarkdownRenderer.tsx`
  - 使用 `react-markdown` 渲染Markdown内容
  - 支持 GFM（表格、任务列表等）
  - 支持代码高亮（可选，使用 `react-syntax-highlighter`）
  - 样式：使用Tailwind CSS定制Markdown样式

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 组件依赖库，配置简单
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (与Task 16-18, 20并行)
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 32 (对话详情页Markdown)
  - **Blocked By**: Task 1

  **References**:
  - Pattern: `frontend/src/pages/TeacherKnowledgeChatDetailPage.tsx` - 现有Markdown渲染
  - External: `https://github.com/remarkjs/react-markdown` - react-markdown文档

  **Acceptance Criteria**:
  - [ ] 组件可渲染Markdown文本
  - [ ] 支持GFM（表格、任务列表）
  - [ ] 代码块正确渲染

- [x] 20. **SSE消息组件（流式显示）**

  **What to do**:
  - 创建 `src/components/ChatMessage.tsx`
  - 实现流式文本渲染（逐字显示效果）
  - 实现引用块展示（折叠/展开）
  - 实现消息历史显示（用户消息 + AI回复）
  - 使用MarkdownRenderer渲染AI回复
  - 样式：用户消息右对齐，AI消息左对齐

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: 流式显示需要动画，引用UI需要设计
  - **Skills**: [`frontend-design`]

  **Parallelization**:
  - **Can Run In Parallel**: YES (与Task 16-19并行)
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 32 (对话详情页)
  - **Blocked By**: Task 12, Task 15, Task 19

  **References**:
  - Pattern: `frontend/src/pages/TeacherKnowledgeChatDetailPage.tsx` - 现有对话消息UI

  **Acceptance Criteria**:
  - [ ] 流式文本逐字显示
  - [ ] 引用块可展开/折叠
  - [ ] 消息历史正确显示（用户右，AI左）
  - [ ] Markdown内容正确渲染

  **QA Scenarios**:
  ```bash
  Scenario: Stream message displays incrementally
    Tool: Bash
    Preconditions: Component created
    Steps:
      1. Pass stream tokens to component one-by-one
      2. Verify text appears gradually
      3. Verify no flickering or jumping
    Expected Result: Smooth streaming display
    Evidence: .sisyphus/evidence/task-20-stream-display.png
  ```

**QA Scenarios**:
  ```bash
  Scenario: Stream message displays incrementally
    Tool: Bash
    Preconditions: Component created
    Steps:
      1. Pass stream tokens to component one-by-one
      2. Verify text appears gradually
      3. Verify no flickering or jumping
    Expected Result: Smooth streaming display
    Evidence: .sisyphus/evidence/task-20-stream-display.png
  ```

- [x] 21. **登录/注册页面（AuthPage）**
  - **What to do**: 实现登录和注册表单，支持email/username登录，role选择（teacher/student），调用authService，成功后跳转到对应布局
  - **Recommended Agent**: `visual-engineering`
  - **Blocks**: 登录流程才能继续后续页面测试
  - **References**: `frontend/src/pages/AuthPage.tsx`, API: `/api/v1/auth/login`, `/register`
  - **QA**: 登录成功跳转、注册成功跳转、错误提示、Token存储

- [x] 22. **课程列表页（TeacherCoursesPage）**
  - **What to do**: 实现课程列表、搜索、筛选、分页、创建/编辑/删除课程，调用courseService
  - **Recommended Agent**: `visual-engineering`
  - **References**: `frontend/src/pages/TeacherCoursesPage.tsx`, API: `/api/v1/teacher/courses`
  - **QA**: 课程列表显示、搜索筛选、创建课程、编辑课程、删除课程

- [x] 23. **课程创建/编辑** (已在Task 22中实现)

- [x] 24. **章节管理（Chapter CRUD）** (已在TeacherCourseContentPage中实现)

- [x] 25. **节管理（Section CRUD）** (已在TeacherCourseContentPage中实现)

- [x] 26. **资源管理（Resource CRUD）** (已在TeacherCourseContentPage中实现)

---

### Wave 4: Teacher Knowledge + Chat (Tasks 27-32)

- [x] 27. **资源预览页（PDF/Video/Audio）**
  - **What to do**: 实现PDF预览（使用react-pdf）、视频预览（HTML video）、音频预览（HTML audio），支持全屏、缩放
  - **Recommended Agent**: `visual-engineering`
  - **References**: `frontend/src/pages/TeacherResourcePreviewPage.tsx`
  - **QA**: PDF正常渲染、视频播放、音频播放

- [x] 28. **知识库入口页（TeacherKnowledgeHubPage）**
  - **What to do**: 显示课程列表，点击课程跳转到知识库资源列表
  - **Recommended Agent**: `visual-engineering`
  - **References**: `frontend/src/pages/TeacherKnowledgeHubPage.tsx`
  - **QA**: 课程列表显示、跳转正确

- [x] 29. **知识库资源列表（TeacherKnowledgeResourcesPage）**
  - **What to do**: 显示课程下的资源列表，显示资源状态（已分块/未分块/已嵌入），点击进入分块管理
  - **Recommended Agent**: `visual-engineering`
  - **References**: `frontend/src/pages/TeacherKnowledgeResourcesPage.tsx`, API: `/api/v1/teacher/courses/:id/knowledge/resources`
  - **QA**: 资源列表显示、状态正确、跳转分块管理

- [x] 30. **知识库分块管理（TeacherKnowledgeChunkPage）**
  - **What to do**: 实现分块预览（设置chunk_size、overlap）、编辑分块内容、保存分块、确认分块、嵌入向量、删除分块
  - **Recommended Agent**: `unspecified-high`
  - **References**: `frontend/src/pages/TeacherKnowledgeChunkPage.tsx`, API: `/api/v1/teacher/resources/:id/knowledge/chunks`
  - **QA**: 分块预览、编辑保存、确认分块、嵌入触发、分块删除

- [x] 31. **对话列表页（TeacherKnowledgeChatsPage）**
  - **What to do**: 显示对话会话列表、创建新会话、删除会话、点击进入对话详情
  - **Recommended Agent**: `visual-engineering`
  - **References**: `frontend/src/pages/TeacherKnowledgeChatsPage.tsx`, API: `/api/v1/teacher/knowledge/chats/sessions`
  - **QA**: 会话列表显示、创建会话、删除会话

- [x] 32. **对话详情页（SSE流式）**
  - **What to do**: 实现对话界面、输入框、发送按钮、调用chatService.askInSessionStream显示流式回复、显示引用块、参数调节（TopK、阈值）
  - **Recommended Agent**: `deep`
  - **References**: `frontend/src/pages/TeacherKnowledgeChatDetailPage.tsx`, API: `/api/v1/teacher/knowledge/chats/sessions/:id/ask/stream`
  - **QA**: 流式回复显示、引用块展开、TopK调节、历史消息加载

---

### Wave 5: Teacher AI Models + Student Pages (Tasks 33-37)

- [x] 33. **AI模型管理页（TeacherAIModelsPage）**
  - **What to do**: 显示AI模型列表（QA/Embedding/Rerank三类）、创建模型、编辑模型、测试连接、删除模型
  - **Recommended Agent**: `visual-engineering`
  - **References**: `frontend/src/pages/TeacherAIModelsPage.tsx`, API: `/api/v1/teacher/ai-models`
  - **QA**: 模型列表显示、创建模型、测试连接、删除模型

- [x] 34. **学生端-课程列表页（StudentCoursesPage）**
  - **What to do**: 显示已发布课程列表、搜索、筛选、选课按钮、点击查看课程目录
  - **Recommended Agent**: `visual-engineering`
  - **References**: 现有教师端课程列表参考, API: `/api/v1/student/courses`, `/courses/:id/enroll`
  - **QA**: 课程列表显示、选课成功、跳转课程目录

- [x] 35. **学生端-我的课程页（StudentMyCoursesPage）**
  - **What to do**: 显示已选课程列表、进度百分比、继续学习按钮
  - **Recommended Agent**: `visual-engineering`
  - **References**: API: `/api/v1/student/my/courses`
  - **QA**: 课程列表显示、进度显示、继续学习跳转

- [x] 36. **学生端-课程目录页（StudentCourseCatalogPage）**
  - **What to do**: 显示课程的章节→节→资源层级结构、点击资源进入学习进度页、显示已完成标记
  - **Recommended Agent**: `visual-engineering`
  - **References**: API: `/api/v1/student/my/courses/:id/catalog`
  - **QA**: 目录结构显示、资源点击跳转、已完成标记

- [x] 37. **学生端-学习进度页（StudentProgressPage）**
  - **What to do**: 显示资源内容（视频/PDF）、进度条、观看时长、标记完成按钮、调用进度上报API
  - **Recommended Agent**: `visual-engineering`
  - **References**: API: `/api/v1/student/my/resources/:id/progress`, `/complete`
  - **QA**: 视频播放、进度上报、标记完成

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.

- [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, curl endpoint, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high`
  Run `npm run build` + `npm run lint` + `npm run test`. Review all changed files for: `as any`/`@ts-ignore`, empty catches, console.log in prod, commented-out code, unused imports. Check AI slop: excessive comments, over-abstraction, generic names.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [x] F3. **Real Manual QA** — `unspecified-high` (+ `playwright` skill if UI)
  Start from clean state. Execute EVERY QA scenario from EVERY task — follow exact steps, capture evidence. Test cross-task integration (features working together, not isolation). Test edge cases: empty state, invalid input, rapid actions. Save to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (git log/diff). Verify 1:1 — everything in spec was built (no missing), nothing beyond spec was built (no creep). Check "Must NOT do" compliance. Detect cross-task contamination: Task N touching Task M's files. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

- **Wave 1**: `chore(init): project setup and configuration` - package.json, vite.config.ts, tsconfig.json, tailwind.config.js
- **Wave 2**: `feat(services): API service layer and auth store` - src/services/, src/stores/
- **Wave 3**: `feat(ui): shared components and layouts` - src/components/, src/layouts/
- **Wave 4**: `feat(teacher): auth and course management pages` - src/pages/teacher/AuthPage.tsx, etc.
- **Wave 5**: `feat(teacher): knowledge and chat pages` - src/pages/teacher/knowledge/, src/pages/teacher/chats/
- **Wave 6**: `feat(all): AI models and student pages` - src/pages/teacher/TeacherAIModelsPage.tsx, src/pages/student/
- **Final**: `test(all): add integration tests` - src/**/*.test.ts(x)

---

## Success Criteria

### Verification Commands
```bash
npm run dev          # Dev server starts on http://localhost:5173
npm run build        # Production build succeeds
npm run test         # Tests pass, coverage >= 60%
npm run lint         # No linting errors
curl http://localhost:5173/health  # Health check (if implemented)
```

### Final Checklist
- [ ] All "Must Have" implemented
- [ ] All "Must NOT Have" absent
- [ ] All tests pass
- [ ] Teacher端10页完整
- [ ] 学生端4页完整
- [ ] SSE流式对话工作