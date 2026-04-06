package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	// Port 服务监听端口。
	Port string `mapstructure:"port"`
}

// PostgresConfig 使用结构化字段描述连接信息，避免把连接串冗余在一个 dsn 字符串里。
type PostgresConfig struct {
	// Host PostgreSQL 主机地址。
	Host string `mapstructure:"host"`
	// Port PostgreSQL 端口。
	Port int `mapstructure:"port"`
	// User PostgreSQL 用户名。
	User string `mapstructure:"user"`
	// Password PostgreSQL 密码。
	Password string `mapstructure:"password"`
	// Database PostgreSQL 数据库名。
	Database string `mapstructure:"database"`
	// SSLMode PostgreSQL SSL 模式（disable/require 等）。
	SSLMode string `mapstructure:"sslmode"`
	// MaxOpenConns 连接池最大打开连接数。
	MaxOpenConns int `mapstructure:"max_open_conns"`
	// MaxIdleConns 连接池最大空闲连接数。
	MaxIdleConns int `mapstructure:"max_idle_conns"`
	// ConnMaxIdleMinutes 连接可空闲的最大分钟数。
	ConnMaxIdleMinutes int `mapstructure:"conn_max_idle_min"`
	// ConnMaxLifetimeMinutes 连接可复用的最大生命周期分钟数。
	ConnMaxLifetimeMinutes int `mapstructure:"conn_max_lifetime_min"`
}

// DSN 根据结构化字段动态拼接 postgres 连接串。
func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User,
		p.Password,
		p.Host,
		p.Port,
		p.Database,
		p.SSLMode,
	)
}

type RedisConfig struct {
	// Addr Redis 地址，格式通常为 host:port。
	Addr string `mapstructure:"addr"`
	// DB Redis 逻辑库编号。
	DB int `mapstructure:"db"`
}

type JWTConfig struct {
	// Issuer JWT 签发方标识。
	Issuer string `mapstructure:"issuer"`
	// AccessSecret Access Token 签名密钥。
	AccessSecret string `mapstructure:"access_secret"`
	// RefreshSecret Refresh Token 签名密钥。
	RefreshSecret string `mapstructure:"refresh_secret"`
	// AccessTTLMinutes Access Token 过期时间（分钟）。
	AccessTTLMinutes int `mapstructure:"access_ttl_min"`
	// RefreshTTLMinutes Refresh Token 过期时间（分钟）。
	RefreshTTLMinutes int `mapstructure:"refresh_ttl_min"`
}

// LoggingConfig 应用日志（logrus）相关配置。
type LoggingConfig struct {
	// Level 日志级别：trace / debug / info / warn|warning / error / fatal / panic（不区分大小写，非法值回退为 info）。
	Level string `mapstructure:"level"`
}

type MinioConfig struct {
	// Endpoint MinIO 服务地址，格式通常为 host:port。
	Endpoint string `mapstructure:"endpoint"`
	// AccessKey MinIO AccessKey。
	AccessKey string `mapstructure:"access_key"`
	// SecretKey MinIO SecretKey。
	SecretKey string `mapstructure:"secret_key"`
	// UseSSL 是否启用 HTTPS 连接 MinIO。
	UseSSL bool `mapstructure:"use_ssl"`
	// Bucket 课程资源存储桶名称。
	Bucket string `mapstructure:"bucket"`
	// PublicBaseURL 写入 object_url 时使用的浏览器可达基址（含 scheme，无尾斜杠），例如 http://localhost:9000。
	// 留空则使用 use_ssl + endpoint 拼接，与 SDK 连接地址一致。
	PublicBaseURL string `mapstructure:"public_base_url"`
	// PublicRead 为 true 时自动设置桶策略允许匿名 s3:GetObject（仅适合开发/可信内网；生产应 false 并改用预签名或网关鉴权）。
	PublicRead bool `mapstructure:"public_read"`
	// ConverterImage Office 文档转 PDF 的转换镜像（需要 docker 环境）。
	ConverterImage string `mapstructure:"converter_image"`
}

// DocReaderConfig 多模态解析网关（docreader-http）HTTP 客户端参数。
type DocReaderConfig struct {
	// BaseURL 解析服务根地址，无尾斜杠，例如 http://localhost:8090；留空则教师解析接口返回未配置错误。
	BaseURL string `mapstructure:"base_url"`
	// InternalToken 与网关 X-Internal-Token 一致的共享密钥；留空则不发送该头。
	InternalToken string `mapstructure:"internal_token"`
	// TimeoutSeconds 单次 /v1/read 调用的 HTTP 超时（秒），0 表示使用客户端默认 600s。
	TimeoutSeconds int `mapstructure:"timeout_seconds"`
	// UseOCR 为 true 时请求体带 use_ocr，强制网关对 PDF/图片做 OCR（仍依赖网关 OCR_BACKEND 等环境变量）。
	UseOCR bool `mapstructure:"use_ocr"`
}

// RAGConfig 知识库问答提示词与兜底文案配置。
type RAGConfig struct {
	// QASystemPrompt 问答系统提示词；会注入“引用上下文块 + 用户问题”。
	QASystemPrompt string `mapstructure:"qa_system_prompt"`
	// FallbackAnswer 当上下文不足时要求模型返回的兜底文案。
	FallbackAnswer string `mapstructure:"fallback_answer"`
	// IntentSystemPrompt 意图识别提示词，仅输出 simple 或 rag。
	IntentSystemPrompt string `mapstructure:"intent_system_prompt"`
	// RewriteSystemPrompt RAG 查询改写提示词。
	RewriteSystemPrompt string `mapstructure:"rewrite_system_prompt"`
	// KeywordSystemPrompt 对改写问题做关键词分解的提示词。
	KeywordSystemPrompt string `mapstructure:"keyword_system_prompt"`
	// SimpleQASystemPrompt 简单问答路径的系统提示词。
	SimpleQASystemPrompt string `mapstructure:"simple_qa_system_prompt"`
}

// Config 按模块分层组织配置，避免所有字段平铺在一个层级。
type Config struct {
	// Server 服务层配置。
	Server ServerConfig `mapstructure:"server"`
	// Postgres 关系型数据库配置。
	Postgres PostgresConfig `mapstructure:"postgres"`
	// Redis 缓存与会话存储配置。
	Redis RedisConfig `mapstructure:"redis"`
	// JWT 鉴权令牌配置。
	JWT JWTConfig `mapstructure:"jwt"`
	// Logging 日志级别等。
	Logging LoggingConfig `mapstructure:"logging"`
	// Minio 对象存储配置。
	Minio MinioConfig `mapstructure:"minio"`
	// DocReader 文档解析 HTTP 网关配置。
	DocReader DocReaderConfig `mapstructure:"docreader"`
	// RAG 知识库问答提示词配置。
	RAG RAGConfig `mapstructure:"rag"`
}

// Load 使用 viper 统一加载配置，覆盖顺序如下（后者覆盖前者）：
// 1) config/config.yaml
// 2) 系统环境变量（需配合 viper.AutomaticEnv）
// .env 由启动命令在 shell 层 source 后注入系统环境变量。
func Load() Config {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	// 保留根目录作为兜底路径，便于兼容历史目录结构。
	v.AddConfigPath(".")
	_ = v.ReadInConfig()

	// 环境变量覆盖：例如 SERVER_PORT -> server.port。
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 默认值兜底，避免缺配置时启动失败。
	v.SetDefault("server.port", "8080")
	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", 5432)
	v.SetDefault("postgres.user", "postgres")
	v.SetDefault("postgres.password", "postgres")
	v.SetDefault("postgres.database", "online_course")
	v.SetDefault("postgres.sslmode", "disable")
	v.SetDefault("postgres.max_open_conns", 30)
	v.SetDefault("postgres.max_idle_conns", 10)
	v.SetDefault("postgres.conn_max_idle_min", 15)
	v.SetDefault("postgres.conn_max_lifetime_min", 60)
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("jwt.issuer", "online-course")
	v.SetDefault("jwt.access_secret", "dev-access-secret")
	v.SetDefault("jwt.refresh_secret", "dev-refresh-secret")
	v.SetDefault("jwt.access_ttl_min", 60)
	v.SetDefault("jwt.refresh_ttl_min", 60*24*7)
	v.SetDefault("logging.level", "info")
	v.SetDefault("minio.endpoint", "localhost:9000")
	v.SetDefault("minio.access_key", "minioadmin")
	v.SetDefault("minio.secret_key", "minioadmin")
	v.SetDefault("minio.use_ssl", false)
	v.SetDefault("minio.bucket", "course-resources")
	v.SetDefault("minio.public_base_url", "")
	v.SetDefault("minio.public_read", false)
	v.SetDefault("minio.converter_image", "jrottenberg/libreoffice:7.6")
	v.SetDefault("docreader.base_url", "")
	v.SetDefault("docreader.internal_token", "")
	v.SetDefault("docreader.timeout_seconds", 600)
	v.SetDefault("docreader.use_ocr", false)
	v.SetDefault("rag.qa_system_prompt", `## 角色
你是课程知识库问答助手，只能基于提供的课程检索上下文回答问题。

## 问题
用户会给出当前提问，以及若干编号的检索上下文块。

## 目标
给出准确、简洁、可执行的答案；当证据不足时，不臆造事实。

## 要求
- 仅可使用检索上下文中的事实，不得引入上下文外信息。
- 若上下文不足以回答，必须原样回复：{{fallback_answer}}。
- 不要在正文中输出引用序号（如 [1][2]），也不要粘贴大段原文。
- 若有结论冲突，优先采用更具体、更新、与问题更直接相关的片段。`)
	v.SetDefault("rag.fallback_answer", "我在当前课程知识库中找不到足够依据，请补充资料或换个问法。")
	v.SetDefault("rag.intent_system_prompt", `## 角色
你是对话意图分类器。

## 问题
输入包含历史对话与当前问题，需要判断本轮应走 simple 还是 rag。

## 目标
输出最稳定、最可执行的路由决策，避免误分流。

## 要求
- 只能输出一个小写词：simple 或 rag，禁止输出其他文本。
- simple：通用常识、闲聊、无需课程知识库即可回答。
- rag：需要课程资料、术语定义、课程细节或证据支撑才能回答。
- 若不确定，默认输出 rag。`)
	v.SetDefault("rag.rewrite_system_prompt", `## 角色
你是检索查询改写器。

## 问题
输入包含历史对话与当前问题，需要产出更适合检索的查询。

## 目标
在不改变用户意图的前提下，提高语义检索与关键词检索召回率。

## 要求
- 只输出一条改写后的查询文本，不要输出解释。
- 保留关键实体、课程术语、约束条件（时间、版本、范围等）。
- 去除无信息量口语与礼貌语，避免过度扩写。
- 如当前问题依赖历史指代，需补全指代对象。`)
	v.SetDefault("rag.keyword_system_prompt", `## 角色
你是检索关键词分解器。

## 问题
输入是一条已改写后的检索问题，需要产出用于关键词检索的分词关键词。

## 目标
输出高召回且低噪音的关键词集合，覆盖核心实体、术语与约束词。

## 要求
- 仅输出关键词列表，不要解释。
- 输出格式必须是 JSON 数组字符串，例如：["goroutine","调度","GPM"]。
- 关键词数量控制在 3~8 个。
- 保留专有名词、英文缩写、数字版本信息。`)
	v.SetDefault("rag.simple_qa_system_prompt", "你是通用问答助手，请直接回答用户问题。")

	// 绑定到结构体。
	var conf Config
	_ = v.Unmarshal(&conf)
	return conf
}

// AccessTTL 返回 Access Token 的有效期。
func (c Config) AccessTTL() time.Duration {
	return time.Duration(c.JWT.AccessTTLMinutes) * time.Minute
}

// RefreshTTL 返回 Refresh Token 的有效期。
func (c Config) RefreshTTL() time.Duration {
	return time.Duration(c.JWT.RefreshTTLMinutes) * time.Minute
}
