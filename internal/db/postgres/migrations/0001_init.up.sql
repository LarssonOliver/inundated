BEGIN;

-- Tags
CREATE TABLE tags (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    color TEXT NOT NULL
);

-- Projects
CREATE TABLE projects (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    color TEXT NOT NULL,
    time_budget INTERVAL
);

-- Project-Tag join table
CREATE TABLE project_tags (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (project_id, tag_id)
);

-- Timespans
CREATE TABLE timespans (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL
);

-- TimeSpan-Tag join table
CREATE TABLE timespan_tags (
    timespan_id UUID NOT NULL REFERENCES timespans(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (timespan_id, tag_id)
);

COMMIT;
