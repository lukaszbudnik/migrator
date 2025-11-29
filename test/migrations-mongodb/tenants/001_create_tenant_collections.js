db.users.createIndex({ email: 1 }, { unique: true });
db.settings.insertOne({ key: "initialized", value: true });
