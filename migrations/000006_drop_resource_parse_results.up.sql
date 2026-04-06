-- 解析结果改由分块/嵌入管线承载，移除独立解析结果表（无外键依赖时可安全 DROP）。
DROP TABLE IF EXISTS resource_parse_results;
