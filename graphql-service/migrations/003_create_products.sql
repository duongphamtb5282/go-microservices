-- +up
-- Create products collection with indexes
db.createCollection("products");

-- Create indexes
db.products.createIndex({ "name": 1 });
db.products.createIndex({ "category": 1 });
db.products.createIndex({ "price": 1 });
db.products.createIndex({ "created_at": 1 });
db.products.createIndex({ "is_active": 1 });

-- +down
-- Drop products collection
db.products.drop();
