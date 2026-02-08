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
    project_id UUID NOT NULL,
    tag_id UUID NOT NULL,
    PRIMARY KEY (project_id, tag_id),
    CONSTRAINT project_tags_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT project_tags_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
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
    timespan_id UUID NOT NULL,
    tag_id UUID NOT NULL,
    PRIMARY KEY (timespan_id, tag_id),
    CONSTRAINT timespan_tags_timespan_id_fkey FOREIGN KEY (timespan_id) REFERENCES timespans(id) ON DELETE CASCADE,
    CONSTRAINT timespan_tags_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

COMMIT;
