CREATE DATABASE IF NOT EXISTS micce-search-engine-db;

CREATE TABLE update_process(
    spot_id CHAR(21) NOT NULL,
    updated_at DATETIME NOT NULL,
    vespa_updated_at DATETIME DEFAULT NULL,
    is_vespa_updated TINYINT DEFAULT 0,
    index_status CHAR(21) NOT NULL DEFAULT 'READY',
    PRIMARY KEY(spot_id),
    INDEX index_1(index_status)
);
