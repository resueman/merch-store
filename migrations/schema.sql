CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    price INT
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password TEXT NOT NULL -- должен быть захеширован
);

CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE,
    balance INT NOT NULL DEFAULT 1000,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    CHECK (balance >= 0)
);

CREATE TYPE operation_type AS ENUM (
    'transfer',
    'purchase'
);

CREATE TABLE operations (
    id SERIAL PRIMARY KEY,
    account_id INT NOT NULL,
    operation_type operation_type,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE TABLE transfer_operations (
    id SERIAL PRIMARY KEY,
    operation_id INT,      
    sender_account_id INT NOT NULL,
    recipient_account_id INT NOT NULL,
    amount INT NOT NULL,
    FOREIGN KEY (operation_id) REFERENCES operations(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    FOREIGN KEY (recipient_account_id) REFERENCES accounts(id) ON DELETE CASCADE
    CHECK (sender_account_id <> recipient_account_id)
    CHECK (amount > 0)
);

CREATE TABLE purchase_operations (
    id SERIAL PRIMARY KEY,
    operation_id INT,
    product_id INT NOT NULL,
    customer_account_id INT NOT NULL,
    quantity INT NOT NULL,
    total_price INT NOT NULL,
    FOREIGN KEY (operation_id) REFERENCES operations(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    FOREIGN KEY (customer_account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    CHECK (total_price >= 0),
    CHECK (quantity > 0)
);

-- CREATE INDEX idx_operations_account_id ON operations (account_id);

-- CREATE INDEX idx_operations_user_id ON operations (user_id);

-- CREATE INDEX idx_operations_operation_type ON operations (operation_type);
