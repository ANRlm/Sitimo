-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION problems_update_tsv() RETURNS trigger AS $$
BEGIN
  NEW.search_tsv :=
      setweight(to_tsvector('simple', coalesce(NEW.code, '')), 'A')
    || setweight(to_tsvector('simple', coalesce(NEW.latex, '')), 'B')
    || setweight(to_tsvector('simple', coalesce(NEW.source, '')), 'C')
    || setweight(to_tsvector('simple', coalesce(NEW.notes, '')), 'D');
  NEW.formula_tsv := to_tsvector('simple', coalesce(NEW.formula_tokens, ''));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS problems_tsv_trigger ON problems;

CREATE TRIGGER problems_tsv_trigger
BEFORE INSERT OR UPDATE ON problems
FOR EACH ROW EXECUTE FUNCTION problems_update_tsv();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS problems_tsv_trigger ON problems;
DROP FUNCTION IF EXISTS problems_update_tsv();
-- +goose StatementEnd
