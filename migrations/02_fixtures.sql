-- Test data fixtures for GraphQL tiny example app

-- Insert sample sellers
INSERT INTO sellers (name, address) VALUES
  ('Tech Store', '123 Main St, New York, NY 10001'),
  ('Home Goods', '456 Broadway, San Francisco, CA 94105'),
  ('Fashion Outlet', '789 Market St, Chicago, IL 60607'),
  ('Gadget World', '101 Tech Ave, Seattle, WA 98101');

-- Insert sample listings
INSERT INTO listings (seller_id, title, description, price) VALUES
  (1, 'Smartphone X', 'Latest smartphone with amazing camera', 799.99),
  (1, 'Laptop Pro', 'Powerful laptop for professionals', 1299.99),
  (2, 'Cozy Blanket', 'Super soft winter blanket', 49.99),
  (2, 'Kitchen Mixer', 'Professional grade kitchen mixer', 299.99),
  (3, 'Designer Jeans', 'Premium denim jeans', 89.99),
  (3, 'Casual Shirt', 'Comfortable everyday shirt', 39.99),
  (4, 'Wireless Earbuds', 'True wireless earbuds with great sound', 129.99),
  (4, 'Smart Watch', 'Fitness tracking smart watch', 249.99);

-- Insert sample purchases
INSERT INTO purchases (listing_id, price, bank_tx_id, delivery_address) VALUES
  (1, 799.99, 'TX123456789', '42 Park Avenue, Boston, MA 02215'),
  (3, 49.99, 'TX223456789', '77 Oak Street, Austin, TX 78701'),
  (5, 89.99, 'TX323456789', '15 Pine Road, Portland, OR 97205'),
  (7, 129.99, 'TX423456789', '33 Lake Drive, Miami, FL 33101');

-- Insert sample deliveries
INSERT INTO deliveries (purchase_id, timestamp, status) VALUES
  (1, NOW() - INTERVAL '3 days', 'packed'),
  (1, NOW() - INTERVAL '2 days', 'out_for_delivery'),
  (1, NOW() - INTERVAL '1 day', 'delivered'),
  (2, NOW() - INTERVAL '4 days', 'packed'),
  (2, NOW() - INTERVAL '3 days', 'out_for_delivery'),
  (2, NOW() - INTERVAL '2 days', 'rescheduled'),
  (2, NOW() - INTERVAL '1 day', 'delivered'),
  (3, NOW() - INTERVAL '2 days', 'packed'),
  (3, NOW() - INTERVAL '1 day', 'out_for_delivery'),
  (4, NOW() - INTERVAL '12 hours', 'packed');