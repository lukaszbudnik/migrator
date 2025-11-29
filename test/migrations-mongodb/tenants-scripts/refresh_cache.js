// This script runs every time for each tenant to refresh cache
db.settings.updateMany(
  {},
  { $set: { cached_at: new Date() } }
);
