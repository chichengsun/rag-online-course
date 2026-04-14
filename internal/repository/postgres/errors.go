package postgres

import "errors"

var (
	// ErrNotFound 表示目标记录不存在或无权访问导致未命中。
	ErrNotFound = errors.New("record not found")
	// ErrNoCourseAccess 表示学生未选课，不允许读取目录。
	ErrNoCourseAccess = errors.New("no access to this course")
	// ErrInvalidID 表示查询结果中的标识字段非法。
	ErrInvalidID = errors.New("invalid id")
)
