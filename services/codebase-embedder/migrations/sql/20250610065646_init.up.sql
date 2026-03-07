-- goctl model pg datasource --dir internal/model  --style go_zero --url   postgres://root:password@127.0.0.1:5432/codebase_indexer?sslmode=disabl --table  --table codebase,sync_history,index_history

-- Project repository table
CREATE TABLE codebase
(
    id             integer      NOT NULL,
    client_id      VARCHAR(255) NOT NULL, -- User client identifier, e.g., MAC address
    user_id        VARCHAR(255) NOT NULL, -- User identifier, e.g., email or phone number
    name           VARCHAR(255) NOT NULL, -- Codebase name
    client_path    TEXT         NOT NULL, -- Local path of the project
    status         VARCHAR(50)  NOT NULL, -- Codebase status: expired, active
    path           TEXT         NOT NULL, -- Codebase path
    file_count     INT          NOT NULL,
    total_size     BIGINT       NOT NULL,
    extra_metadata TEXT,
    created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT
    ON TABLE codebase IS 'Stores basic information about project repositories';
COMMENT
    ON COLUMN codebase.id IS 'Unique identifier for the project repository';
COMMENT
    ON COLUMN codebase.client_id IS 'User client identifier, such as MAC address';
COMMENT
    ON COLUMN codebase.user_id IS 'User identifier, such as email or phone number';
COMMENT
    ON COLUMN codebase.name IS 'Name of the project repository';
COMMENT
    ON COLUMN codebase.client_path IS 'Local path of the project on the user''s machine';
COMMENT
    ON COLUMN codebase.path IS 'Path of the codebase';
COMMENT
    ON COLUMN codebase.file_count IS 'Number of files in the project';
COMMENT
    ON COLUMN codebase.total_size IS 'Total size of the project (in bytes)';
COMMENT
    ON COLUMN codebase.extra_metadata IS 'Additional metadata about the project';
COMMENT
    ON COLUMN codebase.created_at IS 'Time when the record was created';
COMMENT
    ON COLUMN codebase.updated_at IS 'Time when the record was last updated';

CREATE SEQUENCE codebase_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE CACHE 1;

ALTER SEQUENCE codebase_id_seq OWNED BY codebase.id;
ALTER TABLE ONLY codebase
    ALTER COLUMN id SET DEFAULT nextval('codebase_id_seq'::regclass);
ALTER TABLE ONLY codebase
    ADD CONSTRAINT codebase_pkey PRIMARY KEY (id);
-- 唯一索引
CREATE UNIQUE INDEX idx_codebase_client_id_path ON codebase (client_id, client_path);


-- Index building task history table
CREATE TABLE index_history
(
    id                  integer     NOT NULL,
    sync_id             INTEGER     NOT NULL, -- sync_history.id
    codebase_id         INTEGER     not null, -- codebase.id
    codebase_path       TEXT        not null, -- codebase
    codebase_name       TEXT        not null, -- codebase
    total_file_count    INTEGER ,
    total_success_count INTEGER ,
    total_fail_count    INTEGER,
    total_ignore_count  INTEGER,
    task_type           VARCHAR(50) NOT NULL, -- vector, relation
    status              VARCHAR(50) NOT NULL, -- pending, running, success, failed
    progress            float,                -- index job progress
    error_message       TEXT,                 -- failed  message
    start_time          TIMESTAMP,
    end_time            TIMESTAMP,
    created_at          TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT
    ON TABLE index_history IS 'Records the history of index building tasks';
COMMENT
    ON COLUMN index_history.id IS 'Unique identifier for the index task history record';
COMMENT
    ON COLUMN index_history.sync_id IS 'ID of the associated synchronization history record';
COMMENT
    ON COLUMN index_history.codebase_id IS 'ID of the associated project repository';
COMMENT
    ON COLUMN index_history.codebase_path IS 'Path of the project repository';
CoMMENT
    ON COLUMN index_history.codebase_name IS 'name of the project repository';
COMMENT
    ON COLUMN index_history.total_file_count IS 'Total number of files';
COMMENT
    ON COLUMN index_history.total_success_count IS 'Total success number of files';
COMMENT
    ON COLUMN index_history.total_fail_count IS 'Total fail number of files';
COMMENT
    ON COLUMN index_history.total_ignore_count IS 'Total ignore number of files';
COMMENT
    ON COLUMN index_history.task_type IS 'Task type: vector, relation';
COMMENT
    ON COLUMN index_history.status IS 'Task status: pending, running, success, failed';
COMMENT
    ON COLUMN index_history.progress IS 'Task progress (floating point number between 0 and 1)';
COMMENT
    ON COLUMN index_history.error_message IS 'Error message if the task failed';
COMMENT
    ON COLUMN index_history.start_time IS 'Task start time';
COMMENT
    ON COLUMN index_history.end_time IS 'Task end time';
COMMENT
    ON COLUMN index_history.created_at IS 'Time when the record was created';
COMMENT
    ON COLUMN index_history.updated_at IS 'Time when the record was last updated';

CREATE SEQUENCE index_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE CACHE 1;

ALTER SEQUENCE index_history_id_seq OWNED BY index_history.id;
ALTER TABLE ONLY index_history
    ALTER COLUMN id SET DEFAULT nextval('index_history_id_seq'::regclass);
ALTER TABLE ONLY index_history
    ADD CONSTRAINT index_history_pkey PRIMARY KEY (id);