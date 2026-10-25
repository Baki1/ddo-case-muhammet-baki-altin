CREATE TABLE "yn-copies" (
    id SERIAL PRIMARY KEY,
    book_id INTEGER NOT NULL REFERENCES "yn-books"(id) ON DELETE CASCADE,
    barcode VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'available', -- available, borrowed, reserved, lost
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "yn-loans" (
    id SERIAL PRIMARY KEY,
    copy_id INTEGER NOT NULL REFERENCES "yn-copies"(id),
    user_id INTEGER NOT NULL REFERENCES "yn-ej-users"(id),
    borrowed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    due_date TIMESTAMP WITH TIME ZONE NOT NULL,
    returned_at TIMESTAMP WITH TIME ZONE,
    renewal_count INTEGER DEFAULT 0,
    -- Ortak teknik kural: Para (ceza) tutarları kuruş cinsinden tam sayı olmalı (float yasak)
    fine_amount INTEGER DEFAULT 0
);

CREATE UNIQUE INDEX idx_active_loan ON "yn-loans" (copy_id) WHERE returned_at IS NULL;