-- Create schema migrations table
CREATE TABLE IF NOT EXISTS schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL,
    PRIMARY KEY (version)
);

-- Create devices table
CREATE TABLE IF NOT EXISTS devices (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    created_at datetime(3) DEFAULT NULL,
    updated_at datetime(3) DEFAULT NULL,
    deleted_at datetime(3) DEFAULT NULL,
    building_id bigint unsigned DEFAULT NULL,
    device_id varchar(255) NOT NULL,
    device_name varchar(255) NOT NULL,
    device_type varchar(50) DEFAULT NULL,
    device_status varchar(50) DEFAULT NULL,
    last_online_time datetime(3) DEFAULT NULL,
    PRIMARY KEY (id),
    KEY idx_devices_deleted_at (deleted_at),
    KEY idx_devices_building_id (building_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4; 