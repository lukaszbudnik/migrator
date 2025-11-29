db.feature_toggles.createIndex({ name: 1 }, { unique: true });
db.feature_toggles.insertOne({ name: "dark_mode", enabled: true });
