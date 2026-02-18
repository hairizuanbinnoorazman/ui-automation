-- Allow version to be 0 for drafts (currently defaults to 1)
ALTER TABLE test_procedures MODIFY COLUMN version INT UNSIGNED NOT NULL DEFAULT 0;

-- Add composite index for efficient draft queries
CREATE INDEX idx_version_is_latest ON test_procedures(version, is_latest);

-- Add index for procedure chains (parent_id, version)
CREATE INDEX idx_root_lookup ON test_procedures(parent_id, version);
