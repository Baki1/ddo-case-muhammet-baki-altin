CREATE TABLE "yn-reservations" (
    id SERIAL PRIMARY KEY,
    book_id INTEGER NOT NULL REFERENCES "yn-books"(id),
    user_id INTEGER NOT NULL REFERENCES "yn-ej-users"(id),
    status VARCHAR(50) DEFAULT 'waiting', -- waiting, active (3 gün tutulur), completed, canceled
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    notified_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE "yn-inventory_sessions" (
    id SERIAL PRIMARY KEY,
    started_by INTEGER NOT NULL REFERENCES "yn-ej-users"(id),
    status VARCHAR(50) DEFAULT 'in_progress', -- in_progress, completed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE "yn-inventory_items" (
    id SERIAL PRIMARY KEY,
    session_id INTEGER NOT NULL REFERENCES "yn-inventory_sessions"(id),
    copy_id INTEGER REFERENCES "yn-copies"(id),
    scanned_barcode VARCHAR(50) NOT NULL,
    status VARCHAR(50), -- found, missing, wrong_shelf, surplus
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "yn-audit_logs" (
    id SERIAL PRIMARY KEY,
    action_type VARCHAR(100) NOT NULL,
    table_name VARCHAR(100) NOT NULL,
    record_id INTEGER NOT NULL,
    performed_by INTEGER REFERENCES "yn-ej-users"(id),
    old_data JSONB,
    new_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);