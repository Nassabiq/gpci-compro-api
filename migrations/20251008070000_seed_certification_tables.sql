-- +goose Up
WITH brand_seeds AS (
    INSERT INTO public.brand_categories (name, slug)
    VALUES
        ('Building Materials', 'building-materials'),
        ('Home Appliances', 'home-appliances')
    ON CONFLICT (slug) DO NOTHING
    RETURNING id, slug
), brand_data AS (
    SELECT id, slug FROM brand_seeds
    UNION ALL
    SELECT id, slug FROM public.brand_categories
    WHERE slug IN ('building-materials', 'home-appliances')
),
brand_insert AS (
    INSERT INTO public.brands (brand_category_id, name, slug)
    SELECT
        (SELECT id FROM brand_data WHERE slug = 'building-materials'),
        'EcoBuild',
        'ecobuild'
    UNION ALL
    SELECT
        (SELECT id FROM brand_data WHERE slug = 'home-appliances'),
        'GreenAppliance',
        'greenappliance'
    ON CONFLICT (brand_category_id, slug) DO NOTHING
    RETURNING id, slug
),
brand_result AS (
    SELECT id, slug FROM brand_insert
    UNION ALL
    SELECT id, slug FROM public.brands
    WHERE slug IN ('ecobuild', 'greenappliance')
),
company_insert AS (
    INSERT INTO public.companies (name, slug, email, website, phone_number, address, image)
    VALUES
        ('PT Eco Build Indonesia', 'pt-eco-build-indonesia', 'info@ecobuild.id', 'https://ecobuild.id', '+62-21-555-0101', 'Jl. Hijau No. 12, Jakarta', 'images/companies/ecobuild.jpg'),
        ('Green Appliance Nusantara', 'green-appliance-nusantara', 'contact@greenappliance.co.id', 'https://greenappliance.co.id', '+62-21-555-0202', 'Jl. Melati No. 45, Bandung', 'images/companies/greenappliance.jpg')
    ON CONFLICT (slug) DO NOTHING
    RETURNING id, slug
),
company_result AS (
    SELECT id, slug FROM company_insert
    UNION ALL
    SELECT id, slug FROM public.companies
    WHERE slug IN ('pt-eco-build-indonesia', 'green-appliance-nusantara')
),
certification_insert AS (
    INSERT INTO public.certifications (name, image, program_id)
    VALUES
        ('ISO 14024 Eco-Label', 'images/certifications/iso-14024.png', 1),
        ('Energy Star Appliance', 'images/certifications/energy-star.png', 2),
        ('Indoor Air Quality Standard', 'images/certifications/iaq-standard.png', 1)
    ON CONFLICT (name, program_id) DO NOTHING
    RETURNING id, name
),
certification_result AS (
    SELECT id, name FROM certification_insert
    UNION ALL
    SELECT id, name FROM public.certifications
    WHERE name IN ('ISO 14024 Eco-Label', 'Energy Star Appliance', 'Indoor Air Quality Standard')
),
status_lookup AS (
    SELECT id, code
    FROM public.lkp_cert_status
    WHERE code IN ('valid', 'pending')
),
product_insert AS (
    INSERT INTO public.products (
        company_id,
        brand_id,
        program_id,
        name,
        slug,
        features,
        reason,
        tshp,
        images,
        is_active
    )
    VALUES
        (
            (SELECT id FROM company_result WHERE slug = 'pt-eco-build-indonesia'),
            (SELECT id FROM brand_result WHERE slug = 'ecobuild'),
            1,
            'EcoBuild Sustainable Concrete',
            'ecobuild-sustainable-concrete',
            'Concrete mix using recycled aggregates and low-carbon cement.',
            'Reduces CO2 emissions by 40% compared to conventional concrete.',
            '{"compressive_strength": "40 MPa", "thermal_conductivity": "1.4 W/mK"}' :: jsonb,
            '["images/products/concrete-1.jpg", "images/products/concrete-2.jpg"]' :: jsonb,
            TRUE
        ),
        (
            (SELECT id FROM company_result WHERE slug = 'green-appliance-nusantara'),
            (SELECT id FROM brand_result WHERE slug = 'greenappliance'),
            2,
            'GreenAppliance Smart AC',
            'greenappliance-smart-ac',
            'Energy efficient air conditioner with smart climate control.',
            'Uses 30% less power and employs R32 refrigerant with low GWP.',
            '{"cooling_capacity": "9000 BTU", "energy_efficiency": "EER 12"}' :: jsonb,
            '["images/products/ac-1.jpg", "images/products/ac-2.jpg"]' :: jsonb,
            TRUE
        )
    ON CONFLICT (brand_id, slug) DO NOTHING
    RETURNING id, slug, program_id
),
product_result AS (
    SELECT id, slug, program_id FROM product_insert
    UNION ALL
    SELECT id, slug, program_id FROM public.products
    WHERE slug IN ('ecobuild-sustainable-concrete', 'greenappliance-smart-ac')
),
product_cert_insert AS (
    INSERT INTO public.product_has_certification (
        product_id,
        certification_id,
        certificate_no,
        issue_date,
        expiry_date,
        status_id,
        document_file,
        meta_json
    )
    VALUES
        (
            (SELECT id FROM product_result WHERE slug = 'ecobuild-sustainable-concrete'),
            (SELECT id FROM certification_result WHERE name = 'ISO 14024 Eco-Label'),
            'GL-ECB-2024-001',
            CURRENT_DATE - INTERVAL '90 days',
            CURRENT_DATE + INTERVAL '275 days',
            (SELECT id FROM status_lookup WHERE code = 'valid'),
            'documents/certifications/ecobuild-iso-14024.pdf',
            '{"audited_by": "EcoCert International", "renewal_cycle_months": 12}' :: jsonb
        ),
        (
            (SELECT id FROM product_result WHERE slug = 'ecobuild-sustainable-concrete'),
            (SELECT id FROM certification_result WHERE name = 'Indoor Air Quality Standard'),
            'GL-ECB-IAQ-2024-004',
            CURRENT_DATE - INTERVAL '120 days',
            CURRENT_DATE + INTERVAL '245 days',
            (SELECT id FROM status_lookup WHERE code = 'pending'),
            'documents/certifications/ecobuild-iaq.pdf',
            '{"inspection_stage": "Final review"}' :: jsonb
        ),
        (
            (SELECT id FROM product_result WHERE slug = 'greenappliance-smart-ac'),
            (SELECT id FROM certification_result WHERE name = 'Energy Star Appliance'),
            'GT-GRN-2024-019',
            CURRENT_DATE - INTERVAL '200 days',
            CURRENT_DATE + INTERVAL '165 days',
            (SELECT id FROM status_lookup WHERE code = 'valid'),
            'documents/certifications/greenappliance-energy-star.pdf',
            '{"assessed_by": "PT Green Assessors", "score": 92.5}' :: jsonb
        )
    ON CONFLICT (product_id, certification_id) DO UPDATE
    SET
        certificate_no = EXCLUDED.certificate_no,
        issue_date = EXCLUDED.issue_date,
        expiry_date = EXCLUDED.expiry_date,
        status_id = EXCLUDED.status_id,
        document_file = EXCLUDED.document_file,
        meta_json = EXCLUDED.meta_json,
        updated_at = NOW()
    RETURNING 1
)
SELECT 1;

-- +goose Down
DELETE FROM public.product_has_certification
WHERE (product_id, certification_id) IN (
    SELECT p.id, c.id
    FROM public.products p
    JOIN public.certifications c ON TRUE
    WHERE (p.slug = 'ecobuild-sustainable-concrete' AND c.name = 'ISO 14024 Eco-Label')
       OR (p.slug = 'ecobuild-sustainable-concrete' AND c.name = 'Indoor Air Quality Standard')
       OR (p.slug = 'greenappliance-smart-ac' AND c.name = 'Energy Star Appliance')
);

DELETE FROM public.products
WHERE slug IN ('ecobuild-sustainable-concrete', 'greenappliance-smart-ac');

DELETE FROM public.certifications
WHERE name IN ('ISO 14024 Eco-Label', 'Energy Star Appliance', 'Indoor Air Quality Standard');

DELETE FROM public.companies
WHERE slug IN ('pt-eco-build-indonesia', 'green-appliance-nusantara');

DELETE FROM public.brands
WHERE slug IN ('ecobuild', 'greenappliance');

DELETE FROM public.brand_categories
WHERE slug IN ('building-materials', 'home-appliances');
