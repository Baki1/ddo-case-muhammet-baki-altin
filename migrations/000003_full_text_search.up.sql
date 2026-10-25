CREATE EXTENSION IF NOT EXISTS pg_trgm;

ALTER TABLE "yn-books" ADD COLUMN subject VARCHAR(255);

ALTER TABLE "yn-books" ADD COLUMN search_vector tsvector;

CREATE OR REPLACE FUNCTION update_search_vector() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('turkish', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('turkish', coalesce(NEW.subject, '')), 'B') ||
    setweight(to_tsvector('turkish', coalesce(NEW.author, '')), 'C') ||
    setweight(to_tsvector('simple', coalesce(NEW.isbn, '')), 'D');
  RETURN NEW;
END
$$ LANGUAGE plpgsql;


CREATE TRIGGER trigger_update_search_vector
BEFORE INSERT OR UPDATE ON "yn-books"
FOR EACH ROW EXECUTE FUNCTION update_search_vector();


CREATE INDEX idx_books_search ON "yn-books" USING GIN(search_vector);
CREATE INDEX idx_books_trgm ON "yn-books" USING GIN (title gin_trgm_ops);