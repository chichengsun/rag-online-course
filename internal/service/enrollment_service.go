package service

import (
	"context"

	dtoLearning "rag-online-course/internal/dto/learning"
	"rag-online-course/internal/repository/postgres"
)

// EnrollmentService 负责选课对象业务编排。
type EnrollmentService struct {
	courseRepo     *postgres.CourseRepository
	enrollmentRepo *postgres.EnrollmentRepository
}

// NewEnrollmentService 创建选课业务服务。
func NewEnrollmentService(courseRepo *postgres.CourseRepository, enrollmentRepo *postgres.EnrollmentRepository) *EnrollmentService {
	return &EnrollmentService{
		courseRepo:     courseRepo,
		enrollmentRepo: enrollmentRepo,
	}
}

// ListPublishedCourses 获取已发布课程列表（供学生选课）。
func (s *EnrollmentService) ListPublishedCourses(ctx context.Context) (dtoLearning.ListCoursesResp, error) {
	var listCoursesResp dtoLearning.ListCoursesResp
	publishedCourses, err := s.courseRepo.ListPublishedCourses(ctx)
	if err != nil {
		return listCoursesResp, err
	}
	listCoursesResp.Items = publishedCourses
	return listCoursesResp, nil
}

// Enroll 学生选课。
func (s *EnrollmentService) Enroll(ctx context.Context, courseID, studentID int64) error {
	return s.enrollmentRepo.EnrollCourse(ctx, courseID, studentID)
}

// MyCourses 获取学生已选课程列表。
func (s *EnrollmentService) MyCourses(ctx context.Context, studentID int64) (dtoLearning.MyCoursesResp, error) {
	var myCoursesResp dtoLearning.MyCoursesResp
	enrolledCourses, err := s.enrollmentRepo.ListMyCourses(ctx, studentID)
	if err != nil {
		return myCoursesResp, err
	}
	myCoursesResp.Items = enrolledCourses
	return myCoursesResp, nil
}
