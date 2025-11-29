// Create products collection with indexes
db.products.createIndex({ sku: 1 }, { unique: true });
db.products.createIndex({ category: 1, price: 1 });

// Insert sample product
db.products.insertOne({
  sku: "PROD-001",
  name: "Sample Product",
  category: "electronics",
  price: 99.99,
  created: new Date()
});
