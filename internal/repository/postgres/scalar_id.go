package postgres

import (
	"fmt"
	"strconv"
)

// scalarIDToString 将驱动返回的主键标量（string / int64 / float64 等）规范为十进制字符串，供嵌套 SQL 与 JSON 使用。
func scalarIDToString(v any) (string, error) {
	switch x := v.(type) {
	case string:
		return x, nil
	case int64:
		return strconv.FormatInt(x, 10), nil
	case int32:
		return strconv.FormatInt(int64(x), 10), nil
	case int:
		return strconv.FormatInt(int64(x), 10), nil
	case float64:
		return strconv.FormatInt(int64(x), 10), nil
	case []byte:
		return string(x), nil
	default:
		return "", fmt.Errorf("unsupported id type %T", v)
	}
}
