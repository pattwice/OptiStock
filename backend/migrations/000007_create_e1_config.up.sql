CREATE TABLE e1_system_config (
    config_key TEXT PRIMARY KEY,
    config_value TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO e1_system_config (config_key, config_value) VALUES
  ('NEAR_EXPIRY_DAYS_DEFAULT', '30'),
  ('LOW_STOCK_ALERT_ENABLED', 'TRUE'),
  ('NEAR_EXPIRY_ALERT_ENABLED', 'TRUE')
ON CONFLICT (config_key) DO NOTHING;
