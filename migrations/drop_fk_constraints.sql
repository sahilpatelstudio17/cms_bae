-- Drop foreign key constraints on approval_requests that don't allow NULL
ALTER TABLE IF EXISTS approval_requests DROP CONSTRAINT IF EXISTS fk_approval_requests_user CASCADE;
ALTER TABLE IF EXISTS approval_requests DROP CONSTRAINT IF EXISTS fk_approval_requests_employee CASCADE;

-- Verify constraints are removed
SELECT constraint_name 
FROM information_schema.table_constraints 
WHERE table_name = 'approval_requests' 
AND constraint_type = 'FOREIGN KEY';
