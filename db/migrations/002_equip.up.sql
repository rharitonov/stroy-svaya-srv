-- up
CREATE TABLE IF NOT EXISTS equip (
    code TEXT NOT NULL PRIMARY KEY,  
    description TEXT,
    type TEXT DEFAULT 'КОПР' CHECK(type IN ('КОПР')),
    unit_type TEXT,
    unit_model TEXT,               
    unit_weight INTEGER,
    unit_power INTEGER,                    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

CREATE TRIGGER IF NOT EXISTS update_equip_timestamp
AFTER UPDATE ON equip
FOR EACH ROW
BEGIN
    UPDATE equip SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
