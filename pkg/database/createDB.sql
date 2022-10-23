CREATE TABLE IF NOT EXISTS Users (
    id       INT NOT NULL PRIMARY KEY,
    balance  NUMERIC NOT NULL,
    reserved NUMERIC NOT NULL
);

CREATE TABLE IF NOT EXISTS Transactions (
    order_id INT NOT NULL,
    service_id INT NOT NULL,
    user_id INT NOT NULL,
    cost NUMERIC NOT NULL,
    order_status TEXT,
    date TIMESTAMP NOT NULL
);