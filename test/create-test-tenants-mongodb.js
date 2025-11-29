// Switch to migrator database
db = db.getSiblingDB('migrator');

// Create tenants collection and insert test tenants
db.migrator_tenants.insertMany([
  { name: 'abc', created: new Date() },
  { name: 'def', created: new Date() },
  { name: 'xyz', created: new Date() }
]);

// Create unique index on tenant name
db.migrator_tenants.createIndex({ name: 1 }, { unique: true });

print('Test tenants created successfully');
