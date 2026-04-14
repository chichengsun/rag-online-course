import { Outlet, NavLink } from 'react-router-dom'
import { BookOpen, GraduationCap } from 'lucide-react'
import { Button } from '@/components/ui/button'

export function StudentLayout() {
  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container flex h-16 items-center justify-between px-4">
          <div className="flex items-center gap-2">
            <GraduationCap className="h-6 w-6 text-primary" />
            <span className="text-lg font-semibold">学习中心</span>
          </div>

          <nav className="flex items-center gap-1">
            <NavLink to="/student/courses">
              {({ isActive }) => (
                <Button
                  variant={isActive ? 'secondary' : 'ghost'}
                  size="default"
                  className="gap-2"
                >
                  <BookOpen className="h-4 w-4" />
                  <span>课程列表</span>
                </Button>
              )}
            </NavLink>

            <NavLink to="/student/my-courses">
              {({ isActive }) => (
                <Button
                  variant={isActive ? 'secondary' : 'ghost'}
                  size="default"
                  className="gap-2"
                >
                  <GraduationCap className="h-4 w-4" />
                  <span>我的课程</span>
                </Button>
              )}
            </NavLink>
          </nav>

          <div className="flex items-center gap-2">
            <Button variant="ghost" size="icon">
              <GraduationCap className="h-5 w-5" />
            </Button>
          </div>
        </div>
      </header>

      <main className="container px-4 py-6">
        <Outlet />
      </main>
    </div>
  )
}