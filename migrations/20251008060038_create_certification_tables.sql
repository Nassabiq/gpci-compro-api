-- +goose Up
-- +goose StatementBegin
-- =========================================================
-- Core lookups
-- =========================================================
CREATE TABLE IF NOT EXISTS public.lkp_product_program (
    id SMALLINT PRIMARY KEY,
    code VARCHAR(40) UNIQUE NOT NULL,
    -- 'green_label', 'green_toll'
    name VARCHAR(80) NOT NULL
);

INSERT INTO
    public.lkp_product_program (id, code, name)
VALUES
    (1, 'green_label', 'Green Label'),
    (2, 'green_toll', 'Green Toll') ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS public.lkp_cert_status (
    id SMALLINT PRIMARY KEY,
    code VARCHAR(30) UNIQUE NOT NULL,
    -- valid, expired, suspended, pending, revoked
    name VARCHAR(60) NOT NULL
);

INSERT INTO
    public.lkp_cert_status (id, code, name)
VALUES
    (1, 'valid', 'Valid'),
    (2, 'expired', 'Expired'),
    (3, 'suspended', 'Suspended'),
    (4, 'pending', 'Pending'),
    (5, 'revoked', 'Revoked') ON CONFLICT (id) DO NOTHING;

-- =========================================================
-- Brand categories
-- =========================================================
CREATE TABLE IF NOT EXISTS public.brand_categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(180) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_brand_categories_slug UNIQUE (slug)
);

-- =========================================================
-- Brands
-- =========================================================
CREATE TABLE IF NOT EXISTS public.brands (
    id BIGSERIAL PRIMARY KEY,
    brand_category_id BIGINT NOT NULL,
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(180) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_brands_category FOREIGN KEY (brand_category_id) REFERENCES public.brand_categories(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    -- slug unik dalam 1 kategori
    CONSTRAINT uk_brands_category_slug UNIQUE (brand_category_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_brands_category ON public.brands (brand_category_id);

-- =========================================================
-- Companies (single image/url)
-- =========================================================
CREATE TABLE IF NOT EXISTS public.companies (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL,
    email VARCHAR(190),
    website VARCHAR(255),
    phone_number VARCHAR(60),
    address TEXT,
    image TEXT,
    -- single image/url/path
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_companies_slug UNIQUE (slug)
);

CREATE INDEX IF NOT EXISTS idx_companies_email ON public.companies (email);

CREATE INDEX IF NOT EXISTS idx_companies_website ON public.companies (website);

-- =========================================================
-- Certifications (master untuk kedua program)
-- =========================================================
CREATE TABLE IF NOT EXISTS public.certifications (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(180) NOT NULL,
    -- nama jenis sertifikasi (mis. 'Green Label', 'Green Toll', dsb.)
    image TEXT,
    -- url/path dokumen/gambar sertifikat (opsional)
    program_id SMALLINT NOT NULL -- 1=green_label, 2=green_toll
    REFERENCES public.lkp_product_program(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_certifications_name_program UNIQUE (name, program_id)
);

-- =========================================================
-- Products (images: multiple via JSONB)
-- =========================================================
CREATE TABLE IF NOT EXISTS public.products (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL,
    brand_id BIGINT NOT NULL,
    program_id SMALLINT NOT NULL DEFAULT 1 REFERENCES public.lkp_product_program(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(220) NOT NULL,
    features TEXT,
    reason TEXT,
    tshp JSONB NOT NULL DEFAULT '{}' :: jsonb,
    images JSONB NOT NULL DEFAULT '[]' :: jsonb,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_products_company FOREIGN KEY (company_id) REFERENCES public.companies(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_products_brand FOREIGN KEY (brand_id) REFERENCES public.brands(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    -- slug unik dalam 1 brand
    CONSTRAINT uk_products_brand_slug UNIQUE (brand_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_products_company ON public.products (company_id);

CREATE INDEX IF NOT EXISTS idx_products_brand ON public.products (brand_id);

CREATE INDEX IF NOT EXISTS idx_products_program ON public.products (program_id);

-- =========================================================
-- Unified product certifications (untuk Green Label & Green Toll)
-- =========================================================
CREATE TABLE IF NOT EXISTS public.product_has_certification (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL,
    certification_id BIGINT NOT NULL,
    certificate_no VARCHAR(140),
    issue_date DATE,
    expiry_date DATE,
    status_id SMALLINT REFERENCES public.lkp_cert_status(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    document_file TEXT,
    meta_json JSONB NOT NULL DEFAULT '{}' :: jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_phc_product FOREIGN KEY (product_id) REFERENCES public.products(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_phc_cert FOREIGN KEY (certification_id) REFERENCES public.certifications(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT uk_phc_product_cert UNIQUE (product_id, certification_id)
);

CREATE INDEX IF NOT EXISTS idx_phc_product ON public.product_has_certification (product_id);

CREATE INDEX IF NOT EXISTS idx_phc_cert ON public.product_has_certification (certification_id);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
-- Drop trigger & function
DROP TRIGGER IF EXISTS trg_guard_product_cert_program ON public.product_has_certification;

DROP FUNCTION IF EXISTS public.fn_guard_product_cert_program;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS public.product_has_certification;

DROP TABLE IF EXISTS public.products;

DROP TABLE IF EXISTS public.certifications;

DROP TABLE IF EXISTS public.companies;

DROP TABLE IF EXISTS public.brands;

DROP TABLE IF EXISTS public.brand_categories;

DROP TABLE IF EXISTS public.lkp_cert_status;

DROP TABLE IF EXISTS public.lkp_product_program;

-- +goose StatementEnd