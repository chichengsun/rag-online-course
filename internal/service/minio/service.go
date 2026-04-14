package minio

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"rag-online-course/internal/config"
)

type Service struct {
	client *minio.Client
	bucket string
	// endpoint 与 useSSL 用于 SDK 连接及未配置 PublicBaseURL 时拼接 object_url。
	endpoint string
	useSSL   bool
	// publicBaseURL 非空时作为 object_url 前缀（须与浏览器可达地址一致，可含 scheme）。
	publicBaseURL string
	// publicRead 为 true 时在 EnsureBucket 后应用匿名可读策略（仅适合开发/可信内网）。
	publicRead bool
	// converterImage Office 文档转换为 PDF 的 docker 镜像。
	converterImage string

	// bucketEnsured 表示本进程内已经确保过桶存在（避免每次都 BucketExists/MakeBucket）。
	bucketEnsured bool
	bucketMu      sync.Mutex
}

// New 创建 MinIO 客户端服务。
func New(cfg config.Config) (*Service, error) {
	client, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKey, cfg.Minio.SecretKey, ""),
		Secure: cfg.Minio.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Service{
		client:         client,
		bucket:         cfg.Minio.Bucket,
		endpoint:       cfg.Minio.Endpoint,
		useSSL:         cfg.Minio.UseSSL,
		publicBaseURL:  strings.TrimSpace(cfg.Minio.PublicBaseURL),
		publicRead:     cfg.Minio.PublicRead,
		converterImage: strings.TrimSpace(cfg.Minio.ConverterImage),
	}, nil
}

// ensureBucket 在首次使用对象存储时创建桶（若不存在），并按配置应用匿名读策略。
func (s *Service) ensureBucket(ctx context.Context) error {
	s.bucketMu.Lock()
	defer s.bucketMu.Unlock()

	if s.bucketEnsured {
		return nil
	}
	if err := s.EnsureBucket(ctx); err != nil {
		return err
	}
	s.bucketEnsured = true
	return nil
}

// EnsureBucket 确保存储桶存在，不存在则自动创建；若开启 publicRead 则设置匿名 GetObject。
func (s *Service) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}
	if s.publicRead {
		return s.applyPublicReadPolicy(ctx)
	}
	return nil
}

// applyPublicReadPolicy 将桶内对象设为匿名可读（s3:GetObject），解决私有桶直链 AccessDenied。
func (s *Service) applyPublicReadPolicy(ctx context.Context) error {
	policy := fmt.Sprintf(
		`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`,
		s.bucket,
	)
	return s.client.SetBucketPolicy(ctx, s.bucket, policy)
}

// NewObjectKey 生成对象键，按课程、章节、节分层，便于管理资源。
func (s *Service) NewObjectKey(courseID, chapterID, sectionID, filename string) string {
	return fmt.Sprintf("courses/%s/chapters/%s/sections/%s/%s-%s", courseID, chapterID, sectionID, uuid.NewString(), filename)
}

// PresignedPutURL 生成前端直传对象存储的预签名上传链接。
func (s *Service) PresignedPutURL(ctx context.Context, objectKey string, ttl time.Duration) (*url.URL, error) {
	if err := s.ensureBucket(ctx); err != nil {
		return nil, err
	}
	return s.client.PresignedPutObject(ctx, s.bucket, objectKey, ttl)
}

// PresignedGetURL 生成短期可访问的预签名下载链接。
func (s *Service) PresignedGetURL(ctx context.Context, objectKey string, ttl time.Duration) (*url.URL, error) {
	if err := s.ensureBucket(ctx); err != nil {
		return nil, err
	}
	return s.client.PresignedGetObject(ctx, s.bucket, objectKey, ttl, nil)
}

// PresignedPreviewURL 生成“预览用”短期链接，强制 inline 并指定响应 Content-Type。
// 该链接用于浏览器内嵌查看，避免默认下载行为。
func (s *Service) PresignedPreviewURL(ctx context.Context, objectKey, contentType string, ttl time.Duration) (*url.URL, error) {
	if err := s.ensureBucket(ctx); err != nil {
		return nil, err
	}
	respParams := make(url.Values)
	respParams.Set("response-content-disposition", "inline")
	if strings.TrimSpace(contentType) != "" {
		respParams.Set("response-content-type", contentType)
	}
	return s.client.PresignedGetObject(ctx, s.bucket, objectKey, ttl, respParams)
}

// objectURLBase 返回 object_url 使用的 scheme+host（或完整 public_base_url），无尾斜杠。
func (s *Service) objectURLBase() string {
	if s.publicBaseURL != "" {
		return strings.TrimRight(s.publicBaseURL, "/")
	}
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, s.endpoint)
}

// ObjectURL 根据对象键生成可持久化的对象地址（不包含签名）；路径段经 PathEscape，避免空格等导致直链失效。
func (s *Service) ObjectURL(objectKey string) string {
	base := s.objectURLBase()
	var b strings.Builder
	b.Grow(len(base) + len(s.bucket) + len(objectKey) + 8)
	b.WriteString(base)
	b.WriteByte('/')
	b.WriteString(s.bucket)
	for _, seg := range strings.Split(objectKey, "/") {
		if seg == "" {
			continue
		}
		b.WriteByte('/')
		b.WriteString(url.PathEscape(seg))
	}
	return b.String()
}

// OfficePreviewPDFObjectKey 为 doc/docx/ppt/pptx 生成“预览用 PDF”的对象键。
// 预览文件写入 `previews/<原对象键不含扩展名>.pdf`。
func (s *Service) OfficePreviewPDFObjectKey(objectKey string) string {
	base := path.Base(objectKey)
	ext := path.Ext(base) // 含前导点，如 .docx
	withoutExt := strings.TrimSuffix(objectKey, ext)
	return "previews/" + withoutExt + ".pdf"
}

// EnsureOfficePreviewPDF 确保 Office 文档已存在预览 PDF；若不存在则从原对象下载并使用 LibreOffice 转换。
// 转换过程只用于“预览”，不影响原文件下载。
func (s *Service) EnsureOfficePreviewPDF(ctx context.Context, objectKey string) (string, error) {
	previewKey := s.OfficePreviewPDFObjectKey(objectKey)

	// 先检查是否已存在，避免重复转换。
	_, statErr := s.client.StatObject(ctx, s.bucket, previewKey, minio.StatObjectOptions{})
	if statErr == nil {
		return previewKey, nil
	}

	if err := s.ensureBucket(ctx); err != nil {
		return "", err
	}

	tmpDir, err := os.MkdirTemp("", "roc-minio-office-preview-*")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	base := path.Base(objectKey)
	ext := path.Ext(base)
	// 使用 ASCII 文件名规避容器内命令对中文路径兼容问题。
	localInputName := "source" + ext
	localInputPath := filepath.Join(tmpDir, localInputName)

	// 下载原始 Office 文件到临时目录。
	obj, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return "", err
	}
	defer obj.Close()

	f, err := os.Create(localInputPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, obj); err != nil {
		return "", err
	}

	// 执行 LibreOffice 转换为 PDF。
	// docker run 的方式避免宿主机未安装 LibreOffice。
	// Image 可通过配置 minio.converter_image 覆盖。
	dockerImage := strings.TrimSpace(s.converterImage)
	if dockerImage == "" {
		dockerImage = "jrottenberg/libreoffice:7.6"
	}

	// jrottenberg/libreoffice 容器中以 /data 作为工作目录挂载。
	// LibreOffice 会输出到 --outdir 指定目录，文件名保持不变（仅扩展名为 .pdf）。
	cmd := exec.CommandContext(
		ctx,
		"docker",
		"run",
		"--rm",
		"-v",
		tmpDir+":/data",
		dockerImage,
		"--headless",
		"--convert-to",
		"pdf",
		"--outdir",
		"/data",
		"/data/"+localInputName,
	)

	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return "", fmt.Errorf("office convert to pdf failed: %w, output=%s", cmdErr, string(output))
	}

	// 推断输出文件名。
	withoutExt := strings.TrimSuffix(localInputName, ext)
	localOutputPath := filepath.Join(tmpDir, withoutExt+".pdf")

	outFile, err := os.Open(localOutputPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	stat, err := outFile.Stat()
	if err != nil {
		return "", err
	}

	// 上传预览 PDF 到 MinIO。
	_, putErr := s.client.PutObject(
		ctx,
		s.bucket,
		previewKey,
		outFile,
		stat.Size(),
		minio.PutObjectOptions{
			ContentType: "application/pdf",
		},
	)
	if putErr != nil {
		return "", putErr
	}

	return previewKey, nil
}
