// This script runs every time to update statistics in ref database
db.getSiblingDB('ref').modules.updateMany(
  {},
  { $set: { last_updated: new Date() } }
);
