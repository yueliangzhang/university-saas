-- Fails if any request still has multiple attachments. No attachment is deleted.
ALTER TABLE uni_leave_proofs
    DROP PRIMARY KEY,
    DROP COLUMN proof_index,
    ADD PRIMARY KEY (leave_id);
