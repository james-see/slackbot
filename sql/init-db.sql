DO $$
BEGIN
    IF NOT EXISTS (SELECT * FROM information_schema.tables 
                   WHERE table_schema = 'public' AND table_name = 'test_slack_data') THEN
    CREATE TABLE public.test_slack_data (
        uuid UUID PRIMARY KEY,
        date_added TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        status VARCHAR(255),
        description TEXT
    );
    INSERT INTO test_slack_data (uuid, status, description) VALUES
    ('123e4567-e89b-12d3-a456-426614174000', 'PENDING', 'Sample description 1'),
    ('123e4567-e89b-12d3-a456-426614174001', 'COMPLETED', 'Sample description 2'),
    ('123e4567-e89b-12d3-a456-426614174002', 'PENDING', 'Sample description 3');
    END IF;
END
$$;

