-- Drop instagram_media_insights table

DROP TRIGGER IF EXISTS update_instagram_media_insights_updated_at ON instagram_media_insights;
DROP TABLE IF EXISTS instagram_media_insights;
