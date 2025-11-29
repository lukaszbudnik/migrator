db.system_config.createIndex({ key: 1 }, { unique: true });
db.system_config.insertOne({ key: "version", value: "1.0" });
