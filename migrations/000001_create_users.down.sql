-- 回滚用户表
DROP TABLE IF EXISTS auth_tokens;
DROP TABLE IF EXISTS users;

DROP FUNCTION IF EXISTS update_updated_at_column();
