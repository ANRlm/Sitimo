-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS tags (
  id          text PRIMARY KEY,
  name        text NOT NULL UNIQUE,
  category    text NOT NULL CHECK (category IN ('topic','source','custom')),
  color       text NOT NULL DEFAULT '#64748B',
  description text,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS images (
  id                text PRIMARY KEY,
  filename          text NOT NULL,
  mime              text NOT NULL,
  size_bytes        bigint NOT NULL,
  width             int NOT NULL,
  height            int NOT NULL,
  storage_path      text NOT NULL,
  thumbnail_path    text NOT NULL,
  description       text,
  parent_image_id   text REFERENCES images(id),
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  deleted_at        timestamptz
);

CREATE TABLE IF NOT EXISTS image_tags (
  image_id text NOT NULL REFERENCES images(id) ON DELETE CASCADE,
  tag_id   text NOT NULL REFERENCES tags(id)   ON DELETE CASCADE,
  PRIMARY KEY (image_id, tag_id)
);

CREATE TABLE IF NOT EXISTS problems (
  id                 text PRIMARY KEY,
  code               text NOT NULL UNIQUE,
  latex              text NOT NULL,
  answer_latex       text,
  solution_latex     text,
  problem_type       text NOT NULL CHECK (problem_type IN ('multiple_choice','fill_blank','solve','proof','other')),
  difficulty         text NOT NULL CHECK (difficulty IN ('easy','medium','hard','olympiad')),
  subjective_score   numeric(3,1) CHECK (subjective_score BETWEEN 0 AND 10),
  subject            text,
  grade              text,
  source             text,
  notes              text,
  search_tsv         tsvector,
  formula_tokens     text,
  formula_tsv        tsvector,
  version            int NOT NULL DEFAULT 1,
  created_at         timestamptz NOT NULL DEFAULT now(),
  updated_at         timestamptz NOT NULL DEFAULT now(),
  deleted_at         timestamptz
);

CREATE INDEX IF NOT EXISTS problems_search_idx      ON problems USING GIN (search_tsv);
CREATE INDEX IF NOT EXISTS problems_formula_idx     ON problems USING GIN (formula_tsv);
CREATE INDEX IF NOT EXISTS problems_difficulty_idx  ON problems (difficulty);
CREATE INDEX IF NOT EXISTS problems_subject_grade   ON problems (subject, grade);
CREATE INDEX IF NOT EXISTS problems_updated_idx     ON problems (updated_at DESC);
CREATE INDEX IF NOT EXISTS problems_latex_trgm_idx  ON problems USING GIN (latex gin_trgm_ops);

CREATE TABLE IF NOT EXISTS problem_tags (
  problem_id text NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
  tag_id     text NOT NULL REFERENCES tags(id)     ON DELETE CASCADE,
  PRIMARY KEY (problem_id, tag_id)
);

CREATE TABLE IF NOT EXISTS problem_images (
  problem_id  text NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
  image_id    text NOT NULL REFERENCES images(id),
  order_index int NOT NULL DEFAULT 0,
  PRIMARY KEY (problem_id, image_id)
);

CREATE TABLE IF NOT EXISTS problem_versions (
  id         text PRIMARY KEY,
  problem_id text NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
  version    int NOT NULL,
  snapshot   jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (problem_id, version)
);

CREATE TABLE IF NOT EXISTS papers (
  id            text PRIMARY KEY,
  title         text NOT NULL,
  subtitle      text,
  school_name   text,
  exam_name     text,
  subject       text,
  duration_min  int,
  total_score   numeric(6,1),
  description   text,
  status        text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','completed','review')),
  instructions  text,
  footer_text   text,
  header_json   jsonb NOT NULL DEFAULT '{}'::jsonb,
  layout_json   jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),
  deleted_at    timestamptz
);

CREATE TABLE IF NOT EXISTS paper_items (
  id             text PRIMARY KEY,
  paper_id       text NOT NULL REFERENCES papers(id) ON DELETE CASCADE,
  problem_id     text NOT NULL REFERENCES problems(id),
  order_index    int NOT NULL,
  score          numeric(5,1) NOT NULL DEFAULT 0,
  image_position text CHECK (image_position IN ('inline','below','right')) DEFAULT 'inline',
  UNIQUE (paper_id, order_index)
);

CREATE TABLE IF NOT EXISTS export_jobs (
  id            text PRIMARY KEY,
  paper_id      text NOT NULL REFERENCES papers(id),
  paper_title   text NOT NULL,
  format        text NOT NULL CHECK (format IN ('latex','pdf')),
  variant       text NOT NULL CHECK (variant IN ('student','answer','both')),
  status        text NOT NULL CHECK (status IN ('pending','processing','done','failed')),
  progress      int NOT NULL DEFAULT 0,
  download_path text,
  error_message text,
  created_at    timestamptz NOT NULL DEFAULT now(),
  started_at    timestamptz,
  completed_at  timestamptz
);

CREATE INDEX IF NOT EXISTS export_jobs_status_idx ON export_jobs (status, created_at DESC);

CREATE TABLE IF NOT EXISTS search_history (
  id           text PRIMARY KEY,
  query        text NOT NULL,
  filters      jsonb NOT NULL DEFAULT '{}'::jsonb,
  result_count int NOT NULL DEFAULT 0,
  created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS saved_searches (
  id         text PRIMARY KEY,
  name       text NOT NULL,
  query      text,
  filters    jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS app_settings (
  key        text PRIMARY KEY,
  value      jsonb NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS problem_code_counters (
  year           int PRIMARY KEY,
  current_serial int NOT NULL DEFAULT 0,
  updated_at     timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS problem_code_counters;
DROP TABLE IF EXISTS app_settings;
DROP TABLE IF EXISTS saved_searches;
DROP TABLE IF EXISTS search_history;
DROP TABLE IF EXISTS export_jobs;
DROP TABLE IF EXISTS paper_items;
DROP TABLE IF EXISTS papers;
DROP TABLE IF EXISTS problem_versions;
DROP TABLE IF EXISTS problem_images;
DROP TABLE IF EXISTS problem_tags;
DROP TABLE IF EXISTS problems;
DROP TABLE IF EXISTS image_tags;
DROP TABLE IF EXISTS images;
DROP TABLE IF EXISTS tags;
DROP EXTENSION IF EXISTS pg_trgm;
-- +goose StatementEnd
