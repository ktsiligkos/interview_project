USE xm_companies;

CREATE TABLE `users` (
  `id` bigint PRIMARY KEY,
  `name` varchar(255),
  `email` varchar(255),
  `password_hash` varchar(255)
);

CREATE TABLE companies (
    id CHAR(36) NOT NULL DEFAULT (UUID()),
    name VARCHAR(15) NOT NULL UNIQUE,
    description VARCHAR(3000),
    amount_of_employees INT NOT NULL,
    registered BOOLEAN NOT NULL,
    type ENUM('Corporations', 'NonProfit', 'Cooperative', 'Sole Proprietorship') NOT NULL,
    PRIMARY KEY (id)
);