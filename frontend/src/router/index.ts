import { createRouter, createWebHistory } from 'vue-router'
import AuthPage from '@/pages/AuthPage.vue'
import TeacherLayout from '@/layouts/TeacherLayout.vue'
import TeacherCoursesPage from '@/pages/teacher/TeacherCoursesPage.vue'
import TeacherCourseContentPage from '@/pages/teacher/TeacherCourseContentPage.vue'
import TeacherResourcePreviewPage from '@/pages/teacher/TeacherResourcePreviewPage.vue'
import TeacherKnowledgeHubPage from '@/pages/teacher/TeacherKnowledgeHubPage.vue'
import TeacherKnowledgeResourcesPage from '@/pages/teacher/TeacherKnowledgeResourcesPage.vue'
import TeacherKnowledgeChunkPage from '@/pages/teacher/TeacherKnowledgeChunkPage.vue'
import TeacherAIModelsPage from '@/pages/teacher/TeacherAIModelsPage.vue'
import TeacherKnowledgeChatsPage from '@/pages/teacher/TeacherKnowledgeChatsPage.vue'
import TeacherKnowledgeChatDetailPage from '@/pages/teacher/TeacherKnowledgeChatDetailPage.vue'
import TeacherCourseDesignHubPage from '@/pages/teacher/TeacherCourseDesignHubPage.vue'
import TeacherCourseDesignPage from '@/pages/teacher/TeacherCourseDesignPage.vue'
import TeacherQuestionBankHubPage from '@/pages/teacher/TeacherQuestionBankHubPage.vue'
import TeacherQuestionBankPage from '@/pages/teacher/TeacherQuestionBankPage.vue'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/teacher/courses' },
    { path: '/auth', component: AuthPage },
    {
      path: '/teacher',
      component: TeacherLayout,
      children: [
        { path: 'courses', component: TeacherCoursesPage },
        { path: 'course-content/:courseId?', component: TeacherCourseContentPage },
        { path: 'course-design', component: TeacherCourseDesignHubPage },
        { path: 'course-design/:courseId', component: TeacherCourseDesignPage },
        { path: 'question-bank', component: TeacherQuestionBankHubPage },
        {
          path: 'question-bank/:courseId',
          redirect: (to) => `/teacher/question-bank/${to.params.courseId}/list`,
        },
        { path: 'question-bank/:courseId/list', component: TeacherQuestionBankPage },
        { path: 'question-bank/:courseId/maintain', component: TeacherQuestionBankPage },
        { path: 'knowledge/:courseId/chunk/:resourceId', component: TeacherKnowledgeChunkPage },
        { path: 'knowledge/chats', component: TeacherKnowledgeChatsPage },
        { path: 'knowledge/chats/:sessionId', component: TeacherKnowledgeChatDetailPage },
        { path: 'knowledge', component: TeacherKnowledgeHubPage },
        { path: 'knowledge/:courseId', component: TeacherKnowledgeResourcesPage },
        { path: 'ai-models', component: TeacherAIModelsPage },
        { path: 'resources/preview', component: TeacherResourcePreviewPage },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.path.startsWith('/teacher') && !auth.isAuthed) {
    return '/auth'
  }
  if (to.path === '/auth' && auth.isAuthed) {
    return '/teacher/courses'
  }
  return true
})

export default router
