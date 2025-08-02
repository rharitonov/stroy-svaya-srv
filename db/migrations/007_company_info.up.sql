-- up
CREATE TABLE IF NOT EXISTS company_info (
    code TEXT NOT NULL PRIMARY KEY,  
    name TEXT,                       
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

CREATE TRIGGER IF NOT EXISTS update_company_info_timestamp
AFTER UPDATE ON company_info
FOR EACH ROW
BEGIN
    UPDATE company_info SET updated_at = CURRENT_TIMESTAMP WHERE code = OLD.code;
END;