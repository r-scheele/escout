CREATE TABLE IF NOT EXISTS users (
  id bigserial PRIMARY KEY,
  username varchar NOT NULL UNIQUE,
  password varchar NOT NULL,
  email varchar NOT NULL UNIQUE,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS products (
  id bigserial PRIMARY KEY,
  user_id bigint NOT NULL,
  name varchar NOT NULL,
  link varchar NOT NULL UNIQUE,
  base_price float NOT NULL,
  percentage_change float NOT NULL,
  tracking_frequency integer NOT NULL,
  notification_threshold float NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS price_changes (
  id bigserial PRIMARY KEY,
  product_id bigint NOT NULL,
  price numeric,
  changed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (product_id) REFERENCES products (id)
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE INDEX IF NOT EXISTS idx_products_user_id ON products (user_id);
CREATE INDEX IF NOT EXISTS idx_price_changes_product_id ON price_changes (product_id);
