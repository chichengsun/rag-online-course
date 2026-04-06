package handlers

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"rag-online-course/internal/api/response"
	"rag-online-course/internal/logging"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
)

// 本文件将 PostgreSQL 唯一约束冲突（SQLSTATE 23505）转为 HTTP 409 与可读中文说明。
// 约束名与表结构来源：migrations/000001_init.up.sql（含显式 UNIQUE INDEX 与表内 UNIQUE）。
//
// 说明：GORM 使用 pgx 驱动时，错误链末端通常为 *pgconn.PgError；lib/pq 仍做兼容；最后再对 err.Error() 做字符串兜底。

// respondPostgresUniqueViolation 若为 23505 则写入 409 与业务文案并打 Warn 日志，返回 true 表示调用方无需再返回 500。
func respondPostgresUniqueViolation(c *gin.Context, err error, logOp string) bool {
	status, msg, ok := postgresUniqueViolationResponse(err)
	if !ok {
		return false
	}
	logging.FromContext(c.Request.Context()).WithError(err).WithField("op", logOp).Warn("postgres unique violation")
	response.Error(c, status, msg)
	return true
}

// postgresUniqueViolationResponse 解析唯一冲突，返回 HTTP 状态（固定 409）、用户可见文案、是否命中。
func postgresUniqueViolationResponse(err error) (status int, message string, ok bool) {
	if err == nil {
		return 0, "", false
	}
	constraint, table, hit := extractPostgresUniqueViolationMeta(err)
	if !hit {
		return 0, "", false
	}
	return http.StatusConflict, userMessageForUniqueViolation(constraint, table), true
}

// extractPostgresUniqueViolationMeta 从错误链或文本中提取约束名与表名（表名仅 pgx 可靠）。
func extractPostgresUniqueViolationMeta(err error) (constraintName, tableName string, ok bool) {
	var pgxErr *pgconn.PgError
	if errors.As(err, &pgxErr) && pgxErr.Code == "23505" {
		return pgxErr.ConstraintName, pgxErr.TableName, true
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return pqErr.Constraint, "", true
	}
	s := err.Error()
	if strings.Contains(s, "23505") && (strings.Contains(s, "duplicate key") || strings.Contains(s, "unique constraint")) {
		return extractUniqueConstraintNameFromText(s), "", true
	}
	return "", "", false
}

var uniqueConstraintNameRe = regexp.MustCompile(`unique constraint "([^"]+)"`)

func extractUniqueConstraintNameFromText(errText string) string {
	m := uniqueConstraintNameRe.FindStringSubmatch(errText)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

// userMessageForUniqueViolation 按约束名优先、表名次之映射中文说明；未知时给出可操作提示。
func userMessageForUniqueViolation(constraint, table string) string {
	switch constraint {
	// —— 显式命名索引（见 migrations） ——
	case "uk_users_username":
		return "该用户名已被注册，请更换用户名后重试"
	case "uk_courses_teacher_title":
		return "您已存在同名课程，请修改课程标题后重试"
	case "uk_course_chapters_course_title":
		return "该课程下已存在同名章节，请修改章节标题后重试"
	case "uk_chapter_resources_course_chapter_title":
		return "该章节下已存在同名资源，请修改资源标题（可与文件名区分）后重试"

	// —— 表内 UNIQUE，PostgreSQL 默认约束名 ——
	case "course_chapters_course_id_sort_order_key":
		return "该课程下章节排序号与已有章节冲突，请更换排序后重试或刷新页面"
	case "chapter_resources_chapter_id_sort_order_key":
		return "该章节下资源排序号已存在，请修改排序后重试或刷新页面"
	case "chapter_resources_object_key_key":
		return "对象存储键重复，请重新发起上传（init-upload）后再确认"
	case "course_enrollments_course_id_student_id_key":
		return "您已选过该课程，无需重复选课"
	case "learning_progress_course_id_student_id_key":
		return "学习进度记录冲突，请稍后重试或联系管理员"
	case "resource_learning_records_resource_id_student_id_key":
		return "该资源的学习记录已存在，无需重复提交"

	default:
		if constraint != "" {
			return "数据与已有记录冲突（约束：" + constraint + "），请修改后重试"
		}
		return userMessageForUniqueViolationByTable(table)
	}
}

// userMessageForUniqueViolationByTable 约束名为空时（例如仅字符串解析到 23505）按表名给通用说明。
func userMessageForUniqueViolationByTable(table string) string {
	switch table {
	case "users":
		return "用户数据冲突（例如用户名已存在），请修改后重试"
	case "courses":
		return "课程数据冲突（例如同教师下标题重复），请修改后重试"
	case "course_chapters":
		return "章节数据冲突（标题或排序重复），请修改后重试"
	case "chapter_resources":
		return "资源数据冲突（标题、排序或对象键重复），请修改后重试"
	case "course_enrollments":
		return "选课数据冲突，您可能已加入该课程"
	case "learning_progress":
		return "学习进度数据冲突，请稍后重试"
	case "resource_learning_records":
		return "资源学习记录冲突，请稍后重试"
	default:
		return "与已有数据冲突，请检查是否重复提交或修改唯一字段后重试"
	}
}
