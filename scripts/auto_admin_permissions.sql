-- 自动为超级管理员分配新权限的触发器
-- 当permissions表新增记录时，自动为超级管理员角色分配该权限

DELIMITER $$

-- 创建触发器：当插入新权限时，自动分配给超级管理员
CREATE TRIGGER auto_assign_permission_to_super_admin
AFTER INSERT ON permissions
FOR EACH ROW
BEGIN
    -- 为超级管理员角色(id=1)在对应租户下自动分配新权限
    INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
    VALUES ('p', 'r1', CONCAT('p', NEW.id), CONCAT('t', NEW.tenant_id), NULL, NULL, NULL);
END$$

-- 创建存储过程：手动为超级管理员补充遗漏的权限
CREATE PROCEDURE assign_missing_permissions_to_super_admin()
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE perm_id BIGINT;
    DECLARE tenant_id BIGINT;
    
    -- 声明游标，查找超级管理员未拥有的权限
    DECLARE permission_cursor CURSOR FOR
        SELECT p.id, p.tenant_id 
        FROM permissions p
        WHERE NOT EXISTS (
            SELECT 1 FROM casbin_rule cr 
            WHERE cr.ptype = 'p' 
            AND cr.v0 = 'r1' 
            AND cr.v1 = CONCAT('p', p.id)
            AND cr.v2 = CONCAT('t', p.tenant_id)
        );
    
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;
    
    OPEN permission_cursor;
    
    permission_loop: LOOP
        FETCH permission_cursor INTO perm_id, tenant_id;
        IF done THEN
            LEAVE permission_loop;
        END IF;
        
        -- 为超级管理员分配权限
        INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
        VALUES ('p', 'r1', CONCAT('p', perm_id), CONCAT('t', tenant_id), NULL, NULL, NULL);
        
    END LOOP;
    
    CLOSE permission_cursor;
    
    -- 返回处理结果
    SELECT CONCAT('已为超级管理员补充 ', ROW_COUNT(), ' 个权限') as result;
END$$

DELIMITER ;

-- 立即执行一次，为现有权限补充分配
CALL assign_missing_permissions_to_super_admin();

-- 验证触发器是否正常工作的测试
-- 可以通过插入测试权限来验证：
/*
INSERT INTO permissions (tenant_id, permission_code, name, description, resource_type, action, status)
VALUES (1, 'test:trigger', '触发器测试权限', '用于测试自动分配权限的触发器', 'feature', 'test', 1);

-- 检查是否自动分配
SELECT * FROM casbin_rule WHERE v1 = CONCAT('p', LAST_INSERT_ID());

-- 清理测试数据
DELETE FROM permissions WHERE permission_code = 'test:trigger';
DELETE FROM casbin_rule WHERE v1 = CONCAT('p', (SELECT id FROM permissions WHERE permission_code = 'test:trigger'));
*/ 