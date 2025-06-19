-- name: GetAds :many
SELECT * FROM ads;

-- name: GetActiveBidsForPlacement :many
SELECT b.id, b.ad_id, b.bid_price, a.title, a.description, a.image_url, a.destination_url, a.cta_label, a.targeting_json
        FROM bids b
        JOIN ads a ON b.ad_id = a.id
        WHERE b.placement_id = $1 AND a.status = 'active';

-- name: InsertAuction :one
INSERT INTO auctions (id, placement_id, request_context) VALUES ($1, $2, $3) RETURNING *;

-- name: InsertAuctionBid :one
INSERT INTO auction_bids (id, auction_id, bid_id, price, is_winner) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: CreateImpression :one
INSERT INTO impressions (id, ad_id, placement_id, auction_id, user_context)
        VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: CreateClick :one
INSERT INTO clicks (id, impression_id) VALUES ($1, $2) RETURNING *;

-- name: CreateCampaign :one
INSERT INTO ad_campaigns (id, advertiser_id, name, budget) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetSessionByToken :one
SELECT advertiser_id FROM sessions WHERE token = $1 AND expires_at > now();

-- name: GetAdvertiserByEmail :one
SELECT * FROM advertisers WHERE email = $1;

-- name: CreateSession :one
INSERT INTO sessions (token, advertiser_id, expires_at) VALUES ($1, $2, $3) RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token = $1;

-- name: CheckCampaignBelongsToAdvertiser :one
SELECT EXISTS (SELECT 1 FROM ad_campaigns WHERE id = $1 AND advertiser_id = $2);

-- name: CreateAd :one
INSERT INTO ads (id, campaign_id, title, description, image_url, destination_url, cta_label, targeting_json, status)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;

-- name: GetCampaignsByAdvertiser :many
SELECT id, name, budget FROM ad_campaigns WHERE advertiser_id = $1;

-- name: GetAdsByCampaign :many
SELECT id, title, description, image_url, destination_url, cta_label FROM ads WHERE campaign_id = $1;

-- name: CountImpressions :one
SELECT COUNT(*) FROM impressions;

-- name: CountClicks :one
SELECT COUNT(*) FROM clicks;

-- name: AverageCTR :one
SELECT
  ROUND(COUNT(c.id)::decimal / NULLIF(COUNT(i.id), 0), 4) AS ctr
FROM impressions i
LEFT JOIN clicks c ON c.impression_id = i.id;

-- name: CountActiveAds :one
SELECT COUNT(*) FROM ads WHERE status = 'active';

-- name: GetLastImpressionByAdID :one
SELECT id FROM impressions WHERE ad_id = $1 ORDER BY timestamp DESC LIMIT 1;

-- name: GetCampaignByID :one
SELECT * FROM ad_campaigns WHERE id = $1;

-- name: UpdateCampaign :one
UPDATE ad_campaigns
SET name = $2, budget = $3, daily_budget = $4, start_date = $5, end_date = $6, status = $7
WHERE id = $1 AND advertiser_id = $8
RETURNING *;

-- name: UpdateCampaignStatus :one
UPDATE ad_campaigns
SET status = $2
WHERE id = $1 AND advertiser_id = $3
RETURNING *;

-- name: GetCampaignMetrics :many
SELECT
    COUNT(DISTINCT i.id) as impressions,
    COUNT(DISTINCT c.id) as clicks,
    CASE
        WHEN COUNT(DISTINCT i.id) > 0
        THEN ROUND(COUNT(DISTINCT c.id)::decimal / COUNT(DISTINCT i.id) * 100, 2)
        ELSE 0
    END as ctr
FROM ad_campaigns ac
LEFT JOIN ads a ON a.campaign_id = ac.id
LEFT JOIN impressions i ON i.ad_id = a.id
LEFT JOIN clicks c ON c.impression_id = i.id
WHERE ac.id = $1
GROUP BY ac.id;

-- name: GetCampaignSpend :one
SELECT COALESCE(SUM(ab.price), 0) as total_spend
FROM ad_campaigns ac
LEFT JOIN ads a ON a.campaign_id = ac.id
LEFT JOIN bids b ON b.ad_id = a.id
LEFT JOIN auction_bids ab ON ab.bid_id = b.id AND ab.is_winner = true
WHERE ac.id = $1;

-- name: GetCampaignPerformanceAnalysis :many
SELECT 
    ac.id,
    ac.name,
    ac.budget,
    COALESCE(SUM(ab.price), 0) as total_spend,
    COUNT(DISTINCT i.id) as impressions,
    COUNT(DISTINCT c.id) as clicks,
    CASE 
        WHEN COUNT(DISTINCT i.id) > 0 
        THEN ROUND(COUNT(DISTINCT c.id)::decimal / COUNT(DISTINCT i.id) * 100, 2)
        ELSE 0 
    END as ctr,
    CASE 
        WHEN COALESCE(SUM(ab.price), 0) > 0 
        THEN ROUND(COUNT(DISTINCT c.id)::decimal / COALESCE(SUM(ab.price), 1), 2)
        ELSE 0 
    END as clicks_per_dollar,
    CASE 
        WHEN ac.budget > 0 
        THEN ROUND((COALESCE(SUM(ab.price), 0)::decimal / ac.budget) * 100, 2)
        ELSE 0 
    END as budget_utilization
FROM ad_campaigns ac
LEFT JOIN ads a ON a.campaign_id = ac.id
LEFT JOIN impressions i ON i.ad_id = a.id
LEFT JOIN clicks c ON c.impression_id = i.id
LEFT JOIN bids b ON b.ad_id = a.id
LEFT JOIN auction_bids ab ON ab.bid_id = b.id AND ab.is_winner = true
WHERE ac.advertiser_id = $1
GROUP BY ac.id, ac.name, ac.budget
ORDER BY clicks_per_dollar DESC;

-- name: GetAdPerformanceAnalysis :many
SELECT 
    a.id,
    a.title,
    ac.name as campaign_name,
    COUNT(DISTINCT i.id) as impressions,
    COUNT(DISTINCT c.id) as clicks,
    CASE 
        WHEN COUNT(DISTINCT i.id) > 0 
        THEN ROUND(COUNT(DISTINCT c.id)::decimal / COUNT(DISTINCT i.id) * 100, 2)
        ELSE 0 
    END as ctr,
    COALESCE(SUM(ab.price), 0) as total_spend,
    CASE 
        WHEN COALESCE(SUM(ab.price), 0) > 0 
        THEN ROUND(COUNT(DISTINCT c.id)::decimal / COALESCE(SUM(ab.price), 1), 2)
        ELSE 0 
    END as clicks_per_dollar
FROM ads a
JOIN ad_campaigns ac ON a.campaign_id = ac.id
LEFT JOIN impressions i ON i.ad_id = a.id
LEFT JOIN clicks c ON c.impression_id = i.id
LEFT JOIN bids b ON b.ad_id = a.id
LEFT JOIN auction_bids ab ON ab.bid_id = b.id AND ab.is_winner = true
WHERE ac.advertiser_id = $1
GROUP BY a.id, a.title, ac.name
ORDER BY clicks_per_dollar DESC;

-- name: GetBudgetAlerts :many
SELECT 
    ac.id,
    ac.name,
    ac.budget,
    COALESCE(SUM(ab.price), 0) as current_spend,
    CASE 
        WHEN ac.budget > 0 
        THEN ROUND((COALESCE(SUM(ab.price), 0)::decimal / ac.budget) * 100, 2)
        ELSE 0 
    END as budget_utilization,
    CASE 
        WHEN ac.daily_budget IS NOT NULL 
        THEN COALESCE(SUM(ab.price), 0) >= ac.daily_budget
        ELSE false 
    END as daily_budget_exceeded
FROM ad_campaigns ac
LEFT JOIN ads a ON a.campaign_id = ac.id
LEFT JOIN bids b ON b.ad_id = a.id
LEFT JOIN auction_bids ab ON ab.bid_id = b.id AND ab.is_winner = true
WHERE ac.advertiser_id = $1 
    AND ac.status = 'active'
GROUP BY ac.id, ac.name, ac.budget, ac.daily_budget
HAVING 
    (ac.budget > 0 AND COALESCE(SUM(ab.price), 0) >= ac.budget * 0.8) OR
    (ac.daily_budget IS NOT NULL AND COALESCE(SUM(ab.price), 0) >= ac.daily_budget * 0.8);

-- name: GetOptimizationRecommendations :many
WITH campaign_performance AS (
    SELECT 
        ac.id,
        ac.name,
        ac.budget,
        COALESCE(SUM(ab.price), 0) as total_spend,
        COUNT(DISTINCT i.id) as impressions,
        COUNT(DISTINCT c.id) as clicks,
        CASE 
            WHEN COUNT(DISTINCT i.id) > 0 
            THEN ROUND(COUNT(DISTINCT c.id)::decimal / COUNT(DISTINCT i.id) * 100, 2)
            ELSE 0 
        END as ctr,
        CASE 
            WHEN COALESCE(SUM(ab.price), 0) > 0 
            THEN ROUND(COUNT(DISTINCT c.id)::decimal / COALESCE(SUM(ab.price), 1), 2)
            ELSE 0 
        END as clicks_per_dollar
    FROM ad_campaigns ac
    LEFT JOIN ads a ON a.campaign_id = ac.id
    LEFT JOIN impressions i ON i.ad_id = a.id
    LEFT JOIN clicks c ON c.impression_id = i.id
    LEFT JOIN bids b ON b.ad_id = a.id
    LEFT JOIN auction_bids ab ON ab.bid_id = b.id AND ab.is_winner = true
    WHERE ac.advertiser_id = $1 AND ac.status = 'active'
    GROUP BY ac.id, ac.name, ac.budget
)
SELECT 
    id,
    name,
    budget,
    total_spend,
    impressions,
    clicks,
    ctr,
    clicks_per_dollar,
    CASE 
        WHEN clicks_per_dollar > 10 THEN 'increase_budget'
        WHEN clicks_per_dollar < 2 THEN 'decrease_budget'
        WHEN ctr < 1.0 THEN 'improve_creative'
        WHEN impressions < 100 THEN 'increase_bid'
        ELSE 'maintain'
    END as recommendation,
    CASE 
        WHEN clicks_per_dollar > 10 THEN 'Esta campaña tiene excelente ROI. Considera aumentar el presupuesto.'
        WHEN clicks_per_dollar < 2 THEN 'Esta campaña tiene bajo ROI. Considera reducir el presupuesto o mejorar el targeting.'
        WHEN ctr < 1.0 THEN 'CTR bajo. Considera mejorar el creativo o el targeting.'
        WHEN impressions < 100 THEN 'Pocas impresiones. Considera aumentar el bid o mejorar el targeting.'
        ELSE 'Rendimiento estable. Mantén la configuración actual.'
    END as recommendation_text
FROM campaign_performance
ORDER BY clicks_per_dollar DESC;