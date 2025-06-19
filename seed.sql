-- Advertisers
INSERT INTO advertisers (id, name, email, balance, created_at) VALUES
  ('00000000-0000-0000-0000-000000000001', 'Acme Corp', 'acme@example.com', 1000.00, now()),
  ('00000000-0000-0000-0000-000000000002', 'Globex Inc', 'globex@example.com', 500.00, now());

-- Campaigns
INSERT INTO ad_campaigns (id, advertiser_id, name, status, budget, daily_budget, start_date, end_date, created_at) VALUES
  ('10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Campaña ACME', 'active', 500.00, 50.00, now(), now() + interval '30 days', now());

-- Ads
INSERT INTO ads (id, campaign_id, title, description, image_url, destination_url, cta_label, targeting_json, status, created_at) VALUES
  ('20000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', '¡Compra ahora!', 'Descuento por tiempo limitado.', 'https://via.placeholder.com/300x250', 'https://acme.com/oferta', 'Ver más', '{"country": "CL"}', 'active', now());

-- Placements
INSERT INTO placements (id, name, description, slot_size, created_at) VALUES
  ('30000000-0000-0000-0000-000000000001', 'Sidebar 300x250', 'Espacio lateral en blog', '300x250', now());

-- Bids
INSERT INTO bids (id, ad_id, placement_id, bid_price, max_daily_spend, created_at) VALUES
  ('40000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', 0.75, 10.00, now());

-- (Opcional) sesión activa para acme@example.com
INSERT INTO sessions (token, advertiser_id, expires_at, created_at) VALUES
  ('50000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', now() + interval '7 days', now());
