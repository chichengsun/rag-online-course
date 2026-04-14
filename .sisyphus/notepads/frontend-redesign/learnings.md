# Frontend Redesign Learnings
## 2026-04-12
- 实现对话服务层 chat.ts：实现会话管理（createSession、getSessions、updateSession、deleteSession、getMessages）以及 SSE 流式问答接口（askInSession、askInSessionStream）。
- 参考知识通道的 SSE 实现，新增基于 fetch 的 SSE 解析，支持 onToken、onReferences、onDone、onError 回调。
- 文件路径：frontend-opencode/src/services/chat.ts（新增）
- 关注点：避免使用 WebSocket，确保 UI 非阻塞，流式数据逐帧处理。
## Task 1: 初始化 Vite + React + TypeScript 项目

### 完成时间
2026-04-12

### 执行步骤
1. 使用 `npm create vite@latest frontend-opencode -- --template react-ts` 初始化项目
2. 安装核心依赖：`npm install react-router-dom zustand`
3. 配置 vite.config.ts：
   - 添加路径别名 `@` → `./src`
   - 添加 API 代理 `/api/v1` → `http://localhost:8080`
4. 配置 tsconfig.app.json：
   - 添加 `baseUrl: "."`
   - 添加 `paths: { "@/*": ["./src/*"] }`
5. 创建 .env.local：`VITE_API_BASE_URL=http://localhost:8080`
6. 验证：开发服务器启动成功 (VITE v8.0.8, http://localhost:5173)

### 关键配置
- **Vite 路径别名**: 使用 `path.resolve(__dirname, './src')` 解析
- **API 代理**: `/api/v1` 精确匹配，target 默认 `http://localhost:8080`
- **TypeScript 路径映射**: 需要同时配置 `baseUrl` 和 `paths`

### 项目结构
```
frontend-opencode/
├── src/
├── public/
├── .env.local          # API_BASE_URL
├── vite.config.ts      # 代理+路径别名
├── tsconfig.app.json   # 路径映射
└── package.json
```

### 端口
- 开发服务器: 5173
- API 后端: 8080

---

## Task 2: 集成 shadcn/ui 和 Tailwind CSS

### 完成时间
2026-04-12

### 执行步骤
1. 安装 Tailwind CSS v3（注意：v4不兼容shadcn/ui）
   - `npm install -D tailwindcss@3 postcss autoprefixer`
   - `npx tailwindcss init -p` 生成配置文件
2. 配置 tailwind.config.js
   - 添加 content paths: `["./index.html", "./src/**/*.{js,ts,jsx,tsx}"]`
   - 添加 shadcn/ui 主题配置（colors, borderRadius, animations）
   - 安装并配置 `tailwindcss-animate` 插件
3. 配置 tsconfig.json
   - 添加 `compilerOptions.baseUrl` 和 `paths` 配置
   - 添加 `ignoreDeprecations: "6.0"` 解决 TypeScript 警告
4. 安装 shadcn/ui 依赖
   - `npm install class-variance-authority clsx tailwind-merge lucide-react`
5. 初始化 shadcn/ui
   - `npx shadcn@latest init --defaults --force`
   - 自动生成 `components.json` 配置文件
   - 自动创建 `src/components/ui/button.tsx` 和 `src/lib/utils.ts`
6. 安装基础组件
   - `npx shadcn@latest add card input textarea select dialog --yes`
   - `npx shadcn@latest add sonner --yes`（toast组件）
7. 配置 src/index.css
   - 添加 `@tailwind base/components/utilities` 指令
   - 添加 shadcn/ui CSS 变量定义（light/dark主题）
   - 添加 `@layer base` 样式（border-border, bg-background等）

### 关键配置
- **Tailwind 版本**: 必须使用 v3，v4 不兼容 shadcn/ui
- **TypeScript 路径别名**: 需要在 tsconfig.json 和 tsconfig.app.json 中都配置
- **CSS 变量**: shadcn/ui 使用 HSL 颜色变量（`hsl(var(--primary))`）
- **组件导入**: 使用 `@/components/ui/button` 路径别名

### 安装的组件
- button（默认已安装）
- card
- input
- textarea
- select
- dialog
- sonner（toast替代品）

### 验证结果
- `npm run build` 构建成功
- 开发服务器启动成功（http://localhost:5173）
- Button 组件可正常渲染（多种变体：default, secondary, destructive, outline, ghost, link）
- Card 组件可正常渲染
# 资源服务层实现完成

## 文件位置
- frontend-opencode/src/services/resource.ts

## 实现的方法
1. initUpload(token, sectionId, data) - 初始化上传，获取 MinIO 预签名 URL
2. confirmResource(token, sectionId, data) - 确认资源入库
3. getResource(token, resourceId) - 获取资源详情
4. getPreviewUrl(token, resourceId) - 获取预览 URL
5. parseResource(token, resourceId) - 解析资源内容
6. summarizeResource(token, resourceId) - 生成 AI 摘要
7. deleteResource(token, resourceId) - 删除资源
8. getSectionResources(token, sectionId) - 获取节下的资源列表
9. reorderResource(token, resourceId, sortOrder) - 调整资源排序
10. updateResource(token, resourceId, title) - 更新资源信息

## 预签名URL上传流程
1. 调用 initUpload 获取预签名 URL (upload_url) 和 object_key
2. 使用 PUT 请求将文件直接上传到 upload_url（在组件层处理）
3. 上传成功后调用 confirmResource 确认资源入库

## 依赖类型
- InitUploadReq, InitUploadResp
- ConfirmResourceReq, ConfirmResourceResp
- Resource
- ParseResourceResp
- SummarizeResourceResp
- PreviewResourceURLResp

## 设计说明
- 使用 src/services/api.ts 中的 request 方法统一处理 HTTP 请求
- 不在 service 层处理文件上传进度（在组件层处理）
- API 路径遵循 /api/v1/teacher/resources/* 和 /api/v1/teacher/sections/:id/resources/* 规范
