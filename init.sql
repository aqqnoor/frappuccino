DO $$
BEGIN
   IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'frappuccino') THEN
      EXECUTE 'CREATE DATABASE frappuccino';
   END IF;
END
$$;

CREATE TYPE order_status AS ENUM ('open', 'closed');
CREATE TYPE unit_types AS ENUM ('ml', 'shots', 'g');

CREATE TABLE menu_items (
    ID SERIAL PRIMARY KEY,
    Name VARCHAR(50) NOT NULL,
    Description TEXT NOT NULL,
    Price NUMERIC(10, 2) NOT NULL CHECK(Price > 0)
);

CREATE TABLE inventory (
    IngredientID SERIAL PRIMARY KEY,
    Name VARCHAR(50) NOT NULL,
    Quantity INT NOT NULL CHECK(Quantity >= 0),
    Unit unit_types NOT NULL
);

CREATE TABLE orders (
    ID SERIAL PRIMARY KEY,
    CustomerName VARCHAR(50) NOT NULL,
    Status order_status DEFAULT 'open',
    Notes JSONB, -- 
    CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_items (
    OrderID INT,
    ProductID INT NOT NULL,
    Quantity INT NOT NULL CHECK(Quantity > 0),
    PRIMARY KEY (OrderID, ProductID),
    FOREIGN KEY (OrderID) REFERENCES orders(ID),
    FOREIGN KEY (ProductID) REFERENCES menu_items(ID)
);

CREATE TABLE price_history (
    HistoryID SERIAL PRIMARY KEY,
    Menu_ItemID INT NOT NULL,
    old_price NUMERIC(10, 2) NOT NULL CHECK(old_price > 0),
    new_price NUMERIC(10, 2) NOT NULL CHECK(new_price > 0),
    ChangedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (Menu_ItemID) REFERENCES menu_items(ID)
);

CREATE TABLE menu_item_ingredients (
    MenuID INT,
    IngredientID INT NOT NULL,
    Quantity INT NOT NULL CHECK(Quantity > 0),
    PRIMARY KEY (MenuID, IngredientID),
    FOREIGN KEY (MenuID) REFERENCES menu_items(ID) ON DELETE CASCADE,
    FOREIGN KEY (IngredientID) REFERENCES inventory(IngredientID)
);

CREATE TABLE order_status_history (
    ID SERIAL PRIMARY KEY,
    OrderID INT NOT NULL,
    OpenedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ClosedAt TIMESTAMP,
    FOREIGN KEY (OrderID) REFERENCES orders(ID)
);

CREATE TABLE inventory_transactions (
    transactionId SERIAL PRIMARY KEY,
    IngredientID INT REFERENCES inventory(IngredientID) ON DELETE CASCADE,
    quantity_change FLOAT NOT NULL,
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- menu_items
CREATE INDEX idx_menu_items_name ON menu_items (Name);

-- inventory
CREATE INDEX idx_inventory_name ON inventory (Name);

-- orders
CREATE INDEX idx_orders_customer_name ON orders (CustomerName);
CREATE INDEX idx_orders_status ON orders (Status);
CREATE INDEX idx_orders_created_at ON orders (CreatedAt);

-- order_items
CREATE INDEX idx_order_items_order_id ON order_items (OrderID);
CREATE INDEX idx_order_items_product_id ON order_items (ProductID);

-- menu_item_ingredients
CREATE INDEX idx_menu_item_ingredients_menu_id ON menu_item_ingredients (MenuID);
CREATE INDEX idx_menu_item_ingredients_ingredient_id ON menu_item_ingredients (IngredientID);

-- search indexes for full text search
CREATE INDEX idx_menu_item_search_id on menu_items using gin(to_tsvector('english' , name || ' ' || COALESCE(description, '')));


-- Функция для логирования изменения цены в price_history
CREATE OR REPLACE FUNCTION log_price_change()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.price <> OLD.price THEN
        INSERT INTO price_history (Menu_ItemID, old_price, new_price, ChangedAt)
        VALUES (OLD.ID,  OLD.price, NEW.price, CURRENT_TIMESTAMP);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER price_change_trigger
AFTER UPDATE ON menu_items
FOR EACH ROW
EXECUTE FUNCTION log_price_change();

-- Функция для логирования изменения статуса заказа
-- Создание функции для триггера при вставке в orders
CREATE OR REPLACE FUNCTION insert_order_status_history()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO order_status_history (OrderID, OpenedAt)
    VALUES (NEW.ID, CURRENT_TIMESTAMP);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER after_insert_orders
AFTER INSERT ON orders
FOR EACH ROW
EXECUTE FUNCTION insert_order_status_history();

-- Функция для логирования изменения статуса заказа
-- Создание функции для триггера при обновлении в orders
CREATE OR REPLACE FUNCTION update_order_status_history()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE order_status_history
    SET ClosedAt = CURRENT_TIMESTAMP
    WHERE OrderID = NEW.ID AND ClosedAt IS NULL;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER after_update_orders
AFTER UPDATE ON orders
FOR EACH ROW
EXECUTE FUNCTION update_order_status_history();


--Автоматическое логирование в inventory_transactions.
CREATE OR REPLACE FUNCTION log_inventory_transaction()
RETURNS TRIGGER AS $$
BEGIN

    IF TG_OP = 'UPDATE' THEN
        IF NEW.quantity <> OLD.quantity THEN
            INSERT INTO inventory_transactions(IngredientID, quantity_change, reason, created_at)
            VALUES (
                OLD.IngredientID,
                NEW.quantity - OLD.quantity,
                'Inventory adjustment',
                CURRENT_TIMESTAMP
            );
        END IF;

    ELSIF TG_OP = 'INSERT' THEN
        INSERT INTO inventory_transactions(IngredientID, quantity_change, reason, created_at)
        VALUES (
            NEW.IngredientID,
            NEW.quantity,
            'Initial stock',
            CURRENT_TIMESTAMP
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER inventory_change_trigger
AFTER INSERT OR UPDATE ON inventory
FOR EACH ROW
EXECUTE FUNCTION log_inventory_transaction();





-- Mock data for menu_items
INSERT INTO menu_items (Name, Description, Price) VALUES
('Caffe Latte', 'Espresso with steamed milk', 3.50),
('Blueberry Muffin', 'Freshly baked muffin with blueberries', 2.00),
('Espresso', 'Strong and bold coffee', 2.50),
('Cappuccino', 'Espresso with steamed milk and foam', 3.00),
('Mocha', 'Espresso with steamed milk and chocolate', 3.75),
('Iced Latte', 'Iced espresso with milk', 3.80),
('Americano', 'Espresso diluted with hot water', 2.80),
('Carrot Cake', 'Delicious spiced cake with cream cheese frosting', 2.50),
('Vanilla Latte', 'Espresso with steamed milk and vanilla syrup', 3.60),
('Chocolate Croissant', 'Flaky croissant with chocolate filling', 2.80);


-- Mock data for inventory
INSERT INTO inventory (Name, Quantity, Unit) VALUES
('Espresso Shot', 500, 'shots'),
('Milk', 5000, 'ml'),
('Flour', 10000, 'g'),
('Blueberries', 2000, 'g'),
('Sugar', 5000, 'g'),
('Butter', 3000, 'g'),
('Chocolate', 1500, 'g'),
('Coffee Beans', 2000, 'g'),
('Cocoa Powder', 1000, 'g'),
('Vanilla Syrup', 800, 'ml');


-- Mock data for menu_item_ingredients
INSERT INTO menu_item_ingredients (MenuID, IngredientID, Quantity) VALUES
(1, 1, 1),  -- Caffe Latte: 1 Espresso Shot
(1, 2, 200),  -- Caffe Latte: 200 ml Milk
(2, 3, 100),  -- Blueberry Muffin: 100 g Flour
(2, 4, 20),  -- Blueberry Muffin: 20 g Butter
(2, 5, 30),  -- Blueberry Muffin: 30 g Sugar
(3, 1, 1),  -- Espresso: 1 Espresso Shot
(4, 1, 1),  -- Cappuccino: 1 Espresso Shot
(4, 2, 200),  -- Cappuccino: 200 ml Milk
(5, 1, 1),  -- Mocha: 1 Espresso Shot
(5, 2, 200),  -- Mocha: 200 ml Milk
(5, 6, 30),  -- Mocha: 30 g Chocolate
(6, 1, 1),  -- Iced Latte: 1 Espresso Shot
(6, 2, 200),  -- Iced Latte: 200 ml Milk
(7, 1, 1),  -- Americano: 1 Espresso Shot
(8, 3, 100),  -- Carrot Cake: 100 g Flour
(8, 4, 20),  -- Carrot Cake: 20 g Butter
(9, 1, 1),  -- Vanilla Latte: 1 Espresso Shot
(9, 2, 200),  -- Vanilla Latte: 200 ml Milk
(10, 7, 50);  -- Chocolate Croissant: 50 g Chocolate


-- Mock data for orders 
--2024
INSERT INTO orders (CustomerName, Status, Notes, CreatedAt) VALUES
('sanzhar', 'open', '{"notes": "No sugar, extra hot"}', '2024-12-01 08:45:00'),
('aqqnoor', 'open', '{"notes": "Double espresso"}', '2024-12-02 09:30:00'),
('brucewayne', 'open', '{"notes": "Extra chocolate syrup"}', '2024-12-03 10:00:00'),
('johndoe', 'open', '{"notes": "No foam, extra strong"}', '2024-12-05 11:00:00'),
('janedoe', 'open', '{"notes": "Add whipped cream"}', '2024-12-06 12:00:00'),
('pparker', 'open', '{"notes": "Light milk foam"}', '2024-12-07 13:30:00'),
('supercustomer', 'open', '{"notes": "Less sugar, extra vanilla syrup"}', '2024-12-10 14:45:00'),
('mmoldabe', 'open', '{"notes": "More coffee, less ice"}', '2024-12-12 16:00:00'),
('tzhakupo', 'open', '{"notes": "Cinnamon topping"}', '2024-12-15 17:30:00'),
('akakimbe', 'open', '{"notes": "Extra traktor"}', '2024-12-17 18:00:00');

-- 2025
INSERT INTO orders (CustomerName, Status, Notes, CreatedAt) VALUES
('Kimberly Blue', 'closed', '{"notes": "Hot and strong"}', '2025-01-02 09:00:00'),
('Liam Green', 'closed', '{"notes": "Cold milk, no sugar"}', '2025-01-04 09:30:00'),
('Megan Black', 'closed', '{"notes": "Extra foam and cinnamon"}', '2025-01-05 10:15:00'),
('Nina Yellow', 'closed', '{"notes": "Extra hot and vanilla syrup"}', '2025-01-06 11:45:00'),
('Oliver White', 'closed', '{"notes": "Less milk, extra coffee"}', '2025-01-07 12:00:00'),
('Peter Red', 'closed', '{"notes": "No whipped cream, add syrup"}', '2025-01-08 13:00:00'),
('Quincy Purple', 'closed', '{"notes": "Iced coffee, extra shot"}', '2025-01-10 14:00:00'),
('Rebecca Grey', 'closed', '{"notes": "Add caramel"}', '2025-01-11 15:30:00'),
('Steve Brown', 'closed', '{"notes": "Add extra ice"}', '2025-01-12 16:45:00'),
('Tina Pink', 'closed', '{"notes": "No milk, extra strong"}', '2025-01-13 17:00:00');



-- 2024
INSERT INTO order_items (OrderID, ProductID, Quantity) VALUES
(1, 1, 1),  -- sanzhar: 1 Caffe Latte
(1, 2, 1),  -- sanzhar: 1 Blueberry Muffin
(2, 1, 2),  -- aqqnoor: 2 Espresso
(3, 5, 1),  -- brucewayne: 1 Mocha
(3, 6, 1),  -- brucewayne: 1 Iced Latte
(4, 1, 1),  -- johndoe: 1 Espresso
(5, 3, 2),  -- janedoe: 2 Carrot Cake
(6, 7, 1),  -- pparker: 1 Americano
(7, 9, 1),  -- supercustomer: 1 Vanilla Latte
(8, 10, 2), -- mmoldabe: 2 Chocolate Croissants
(9, 4, 1),  -- tzhakupo: 1 Cappuccino
(10, 1, 2); -- akakimbe: 2 Espresso

-- 2025
INSERT INTO order_items (OrderID, ProductID, Quantity) VALUES
(11, 2, 1),  -- Kimberly: 1 Blueberry Muffin
(12, 1, 2),  -- Liam: 2 Caffe Latte
(13, 5, 1),  -- Megan: 1 Mocha
(13, 6, 1),  -- Megan: 1 Iced Latte
(14, 1, 1),  -- Nina: 1 Espresso
(15, 7, 1),  -- Oliver: 1 Americano
(16, 8, 1),  -- Peter: 1 Carrot Cake
(17, 10, 2),  -- Quincy: 2 Chocolate Croissants
(18, 2, 1),  -- Rebecca: 1 Blueberry Muffin
(19, 3, 1),  -- Steve: 1 Espresso
(20, 9, 1);  -- Tina: 1 Vanilla Latte
