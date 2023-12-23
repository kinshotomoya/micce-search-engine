CREATE DATABASE IF NOT EXISTS micce_search_engine;

CREATE TABLE update_process(
    spot_id CHAR(21) NOT NULL,
    updated_at DATETIME NOT NULL,
    vespa_updated_at DATETIME DEFAULT NULL,
    is_vespa_updated TINYINT DEFAULT 0,
    PRIMARY KEY(spot_id),
    INDEX index_1(is_vespa_updated) -- ture or falseだが隔たりがありそうなのでindexをはる
);
