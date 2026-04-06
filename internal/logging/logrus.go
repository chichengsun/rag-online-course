package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"rag-online-course/internal/config"

	"github.com/sirupsen/logrus"
)

// projectRoot 为包含 go.mod 的项目根目录，用于将调用栈文件路径显示为相对路径。
var projectRoot string

// Init 根据应用配置初始化全局 logrus：级别、文本格式、完整时间戳、打印调用处相对项目根的路径与行号。
// 须在加载 config.Load() 之后、业务逻辑之前调用一次。
func Init(cfg config.Config) {
	if projectRoot == "" {
		projectRoot = discoverProjectRoot()
	}

	level := logrus.InfoLevel
	if parsed, err := logrus.ParseLevel(strings.TrimSpace(strings.ToLower(cfg.Logging.Level))); err == nil {
		level = parsed
	}
	logrus.SetLevel(level)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&compactColorFormatter{})
}

type compactColorFormatter struct{}

func (f *compactColorFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	levelColor := "\033[32m"
	switch entry.Level {
	case logrus.WarnLevel:
		levelColor = "\033[33m"
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = "\033[31m"
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = "\033[36m"
	}
	reset := "\033[0m"
	keyColor := "\033[36m"
	valueColor := "\033[37m"
	errorValueColor := "\033[31m"
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	level := strings.ToUpper(entry.Level.String())

	caller := ""
	if entry.Caller != nil {
		caller = fmt.Sprintf("%s:%d", fileRelativeToProject(entry.Caller.File), entry.Caller.Line)
	}

	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		valColor := valueColor
		if k == "error" {
			valColor = errorValueColor
		}
		parts = append(parts, fmt.Sprintf("%s%s%s=%s%v%s", keyColor, k, reset, valColor, entry.Data[k], reset))
	}

	line := fmt.Sprintf("%s%s%s[%s]", levelColor, level, reset, timestamp)
	if caller != "" {
		line += " " + caller
	}
	if len(parts) > 0 {
		line += " " + strings.Join(parts, " ")
	}
	if entry.Message != "" {
		line += " " + entry.Message
	}
	line += "\n"
	return []byte(line), nil
}

// discoverProjectRoot 从本包源文件所在目录向上查找 go.mod，得到项目根路径。
func discoverProjectRoot() string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	dir := filepath.Dir(file)
	for {
		modPath := filepath.Join(dir, "go.mod")
		if st, err := os.Stat(modPath); err == nil && !st.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// fileRelativeToProject 将绝对路径转为相对项目根的路径（正斜杠）；无法转换时退回文件名。
func fileRelativeToProject(absFile string) string {
	if projectRoot == "" || absFile == "" {
		return filepath.Base(absFile)
	}
	rel, err := filepath.Rel(projectRoot, absFile)
	if err != nil || strings.HasPrefix(rel, "..") {
		return filepath.Base(absFile)
	}
	return filepath.ToSlash(rel)
}
