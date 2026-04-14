-- 扩展资源类型枚举，支持 Word 文档（与既有 txt/pdf 等并存）。
ALTER TYPE resource_type ADD VALUE IF NOT EXISTS 'doc';
ALTER TYPE resource_type ADD VALUE IF NOT EXISTS 'docx';
