CREATE TABLE goods_logs (
    id Int64,
    project_id Int64,
    name String,
    description String,
    priority Int32,
    removed UInt8,
    event_time DateTime
) ENGINE = MergeTree()
ORDER BY (id, project_id);

