-- up
-- up
CREATE TABLE IF NOT EXISTS item (
    code TEXT NOT NULL PRIMARY KEY,  
    description TEXT,     
    type TEXT DEFAULT 'СВАЯ'
        CHECK(type IN ('СВАЯ')),
    weight INTEGER,                    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

CREATE TRIGGER IF NOT EXISTS update_item_timestamp
AFTER UPDATE ON item
FOR EACH ROW
BEGIN
    UPDATE item SET updated_at = CURRENT_TIMESTAMP WHERE code = OLD.code;
END;
