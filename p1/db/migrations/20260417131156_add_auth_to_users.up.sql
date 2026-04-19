ALTER TABLE users
ADD COLUMN password TEXT NOT NULL DEFAULT '',
ADD COLUMN role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin'));