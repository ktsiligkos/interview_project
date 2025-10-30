USE xm_companies;

-- Insert sample users
INSERT INTO users (id, name, email, password_hash) VALUES
(1, 'John Doe', 'john_doe@example.com', '$2a$12$xmCUjN8kOXZsez0GNLLZquGIJfzCAx3FUmDav5a.afOD0g2tw7OOq');

-- Insert sample companies
INSERT INTO companies (id, name, description, amount_of_employees, registered, type)
VALUES (UUID(), 'XM', 'A leading tech company specializing in trading.', 500, TRUE, 'Corporations');

INSERT INTO companies (id,name, description, amount_of_employees, registered, type)
VALUES (UUID(), 'MSF', 'An international, independent medical humanitarian organization.', 50, TRUE, 'NonProfit');

INSERT INTO companies (id, name, description, amount_of_employees, registered, type)
VALUES (UUID(), 'EcoCoop', 'A cooperative business promoting sustainable farming.', 120, FALSE, 'Cooperative');

INSERT INTO companies (id,name, description, amount_of_employees, registered, type)
VALUES (UUID(), 'QuickFix', 'Small business offering repair services.', 10, TRUE, 'Sole Proprietorship');
