ALTER TABLE uni_leave_proofs
    ADD COLUMN proof_index TINYINT UNSIGNED NOT NULL DEFAULT 1 AFTER leave_id,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (leave_id, proof_index);
