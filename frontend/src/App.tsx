import { Navigate, Route, Routes } from 'react-router-dom'
import type { ReactElement } from 'react'
import { AuthPage } from './pages/AuthPage'
import { TeacherCoursesPage } from './pages/TeacherCoursesPage'
import { TeacherCourseContentPage } from './pages/TeacherCourseContentPage'
import { TeacherResourcePreviewPage } from './pages/TeacherResourcePreviewPage'
import { TeacherKnowledgeHubPage } from './pages/TeacherKnowledgeHubPage'
import { TeacherKnowledgeResourcesPage } from './pages/TeacherKnowledgeResourcesPage'
import { TeacherKnowledgeChunkPage } from './pages/TeacherKnowledgeChunkPage'
import { TeacherKnowledgeChatsPage } from './pages/TeacherKnowledgeChatsPage'
import { TeacherKnowledgeChatDetailPage } from './pages/TeacherKnowledgeChatDetailPage'
import { TeacherAIModelsPage } from './pages/TeacherAIModelsPage'
import { TeacherLayout } from './layouts/TeacherLayout'
import { useAuth } from './store/auth'

function ProtectedRoute({ children }: { children: ReactElement }) {
  const { token } = useAuth()
  if (!token) {
    return <Navigate to="/auth" replace />
  }
  return children
}

function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/teacher/courses" replace />} />
      <Route path="/auth" element={<AuthPage />} />
      <Route
        path="/teacher"
        element={
          <ProtectedRoute>
            <TeacherLayout />
          </ProtectedRoute>
        }
      >
        <Route path="courses" element={<TeacherCoursesPage />} />
        <Route path="course-content" element={<Navigate to="/teacher/courses" replace />} />
        <Route path="course-content/:courseId" element={<TeacherCourseContentPage />} />
        <Route path="knowledge" element={<TeacherKnowledgeHubPage />} />
        <Route path="knowledge/chats" element={<TeacherKnowledgeChatsPage />} />
        <Route path="knowledge/chats/:sessionId" element={<TeacherKnowledgeChatDetailPage />} />
        <Route path="knowledge/:courseId" element={<TeacherKnowledgeResourcesPage />} />
        <Route path="knowledge/:courseId/chunk/:resourceId" element={<TeacherKnowledgeChunkPage />} />
        <Route path="ai-models" element={<TeacherAIModelsPage />} />
        <Route path="resources/preview" element={<TeacherResourcePreviewPage />} />
      </Route>
    </Routes>
  )
}

export default App
