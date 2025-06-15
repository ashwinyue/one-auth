-- 统一修改所有表字段为下划线命名格式
USE miniblog;

-- 修改 user 表
ALTER TABLE user 
    CHANGE COLUMN userID user_id varchar(36) NOT NULL,
    CHANGE COLUMN createdAt created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHANGE COLUMN updatedAt updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;

-- 修改 post 表  
ALTER TABLE post
    CHANGE COLUMN userID user_id varchar(36) NOT NULL,
    CHANGE COLUMN postID post_id varchar(35) NOT NULL,
    CHANGE COLUMN createdAt created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHANGE COLUMN updatedAt updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;

-- 重新创建索引（如果需要）
-- 对于 user 表
DROP INDEX idx_user_userID ON user;
CREATE UNIQUE INDEX idx_user_user_id ON user(user_id);

-- 对于 post 表
DROP INDEX idx_post_postID ON post;  
CREATE UNIQUE INDEX idx_post_post_id ON post(post_id);

-- 显示修改后的表结构
DESCRIBE user;
DESCRIBE post; 