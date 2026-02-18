-- Fix is_latest column: change DEFAULT TRUE to NOT NULL DEFAULT FALSE.
-- The previous DEFAULT TRUE caused drafts (version=0, is_latest=false) to
-- silently get is_latest=1 if the INSERT ever omitted the column.
ALTER TABLE test_procedures
    MODIFY COLUMN is_latest TINYINT(1) NOT NULL DEFAULT 0;
