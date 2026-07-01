-- Optional demo seed data for OptiStock.
-- Migrations already create the admin user and E1 defaults.
-- Run manually after deploy: psql $DATABASE_URL -f scripts/seed.sql

INSERT INTO a1_items (item_code, description, unit, item_type, min_stock_level, expiry_threshold_days, shelf_life_days)
VALUES
  ('RM-DEMO-001', 'Demo Raw Material A', 'KG', 'RM', 100, 30, NULL),
  ('RM-DEMO-002', 'Demo Raw Material B', 'KG', 'RM', 50, 30, NULL),
  ('SFG-DEMO-001', 'Demo Semi-Finished Good', 'EA', 'SFG', 20, 14, 90),
  ('FG-DEMO-001', 'Demo Finished Good', 'EA', 'FG', 10, 7, 180)
ON CONFLICT (item_code) DO NOTHING;

INSERT INTO a2_bom (parent_item_code, component_item_code, qty_per_set, bom_version, is_active)
VALUES
  ('SFG-DEMO-001', 'RM-DEMO-001', 2.5, '1', TRUE),
  ('SFG-DEMO-001', 'RM-DEMO-002', 1.0, '1', TRUE),
  ('FG-DEMO-001', 'SFG-DEMO-001', 1.0, '1', TRUE)
ON CONFLICT DO NOTHING;
