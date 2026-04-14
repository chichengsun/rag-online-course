import { createBrowserRouter, Navigate } from 'react-router-dom'
import type { RouteObject } from 'react-router-dom'

import { AuthPage } from '@/pages/AuthPage'
import { TeacherLayout } from '@/layouts/TeacherLayout'
import { StudentLayout } from '@/layouts/StudentLayout'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { TeacherCoursesPage } from '@/pages/teacher/TeacherCoursesPage'
import { TeacherCourseContentPage } from '@/pages/teacher/TeacherCourseContentPage'
import { TeacherKnowledgeHubPage } from '@/pages/teacher/TeacherKnowledgeHubPage'
import { TeacherKnowledgeResourcesPage } from '@/pages/teacher/TeacherKnowledgeResourcesPage'
import { TeacherKnowledgeChatsPage } from '@/pages/teacher/TeacherKnowledgeChatsPage'
import { TeacherKnowledgeChatDetailPage } from '@/pages/teacher/TeacherKnowledgeChatDetailPage'
import { TeacherKnowledgeChunkPage } from '@/pages/teacher/TeacherKnowledgeChunkPage'
import { TeacherAIModelsPage } from '@/pages/teacher/TeacherAIModelsPage'
import { TeacherResourcePreviewPage } from '@/pages/teacher/TeacherResourcePreviewPage'
import { TeacherCourseDesignHubPage } from '@/pages/teacher/TeacherCourseDesignHubPage'
import { TeacherCourseDesignPage } from '@/pages/teacher/TeacherCourseDesignPage'
import { TeacherQuestionBankHubPage } from '@/pages/teacher/TeacherQuestionBankHubPage'
import { TeacherQuestionBankCourseLayout } from '@/pages/teacher/TeacherQuestionBankCourseLayout'
import { TeacherQuestionBankListPage } from '@/pages/teacher/TeacherQuestionBankListPage'
import { TeacherQuestionBankMaintainPage } from '@/pages/teacher/TeacherQuestionBankMaintainPage'
import { StudentCoursesPage } from '@/pages/student/StudentCoursesPage'
import { StudentMyCoursesPage } from '@/pages/student/StudentMyCoursesPage'
import { StudentProgressPage } from '@/pages/student/StudentProgressPage'
import { StudentCourseCatalogPage } from '@/pages/student/StudentCourseCatalogPage'

const routes: RouteObject[] = [
  {
    path: '/',
    element: <Navigate to="/teacher/courses" replace />,
  },
  {
    path: '/auth',
    element: <AuthPage />,
  },
  {
    path: '/teacher',
    element: (
      <ProtectedRoute>
        <TeacherLayout />
      </ProtectedRoute>
    ),
    children: [
      {
        path: 'courses',
        element: <TeacherCoursesPage />,
      },
      {
        path: 'course-content',
        element: <Navigate to="/teacher/courses" replace />,
      },
      {
        path: 'course-content/:courseId',
        element: <TeacherCourseContentPage />,
      },
      {
        path: 'course-design',
        element: <TeacherCourseDesignHubPage />,
      },
      {
        path: 'course-design/:courseId',
        element: <TeacherCourseDesignPage />,
      },
      {
        path: 'question-bank',
        element: <TeacherQuestionBankHubPage />,
      },
      {
        path: 'question-bank/:courseId',
        element: <TeacherQuestionBankCourseLayout />,
        children: [
          { index: true, element: <Navigate to="list" replace /> },
          { path: 'list', element: <TeacherQuestionBankListPage /> },
          { path: 'maintain', element: <TeacherQuestionBankMaintainPage /> },
        ],
      },
      {
        path: 'knowledge',
        element: <TeacherKnowledgeHubPage />,
      },
      {
        path: 'knowledge/chats',
        element: <TeacherKnowledgeChatsPage />,
      },
      {
        path: 'knowledge/chats/:sessionId',
        element: <TeacherKnowledgeChatDetailPage />,
      },
      {
        path: 'knowledge/:courseId',
        element: <TeacherKnowledgeResourcesPage />,
      },
      {
        path: 'knowledge/:courseId/chunk/:resourceId',
        element: <TeacherKnowledgeChunkPage />,
      },
      {
        path: 'ai-models',
        element: <TeacherAIModelsPage />,
      },
      {
        path: 'resources/preview',
        element: <TeacherResourcePreviewPage />,
      },
    ],
  },
  {
    path: '/student',
    element: (
      <ProtectedRoute>
        <StudentLayout />
      </ProtectedRoute>
    ),
    children: [
      {
        path: 'courses',
        element: <StudentCoursesPage />,
      },
      {
        path: 'my-courses',
        element: <StudentMyCoursesPage />,
      },
      {
        path: 'courses/:courseId/catalog',
        element: <StudentCourseCatalogPage />,
      },
      {
        path: 'courses/:courseId',
        element: <div>Course Learning</div>,
      },
      {
        path: 'resources/:resourceId',
        element: <StudentProgressPage />,
      },
    ],
  },
]

const router = createBrowserRouter(routes)

export { router }
