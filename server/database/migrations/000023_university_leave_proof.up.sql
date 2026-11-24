CREATE TABLE uni_leave_proofs (
    leave_id BIGINT UNSIGNED NOT NULL,
    filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(32) NOT NULL,
    content MEDIUMBLOB NOT NULL,
    PRIMARY KEY (leave_id),
    CONSTRAINT fk_uni_leave_proof_request FOREIGN KEY (leave_id)
        REFERENCES uni_leave_requests(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
