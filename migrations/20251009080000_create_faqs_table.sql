-- +goose Up
CREATE TABLE IF NOT EXISTS public.faqs (
    id BIGSERIAL PRIMARY KEY,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_faqs_deleted_at ON public.faqs (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS public.faqs;

