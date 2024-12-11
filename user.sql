Create TABLE IF NOT EXISTS "contacts" (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES "users"(id), 
    first_name VARCHAR(55) not null,
    last_name VARCHAR(55) not null,
    middle_name VARCHAR(55) not null,
    phone_number VARCHAR(55) UNIQUE not null,
    created_at TIMESTAMP Default now(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "users" (
    id UUID PRIMARY KEY, 
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    phone_number VARCHAR(15) NOT NULL UNIQUE, 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "devices" (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL, 
    device_info VARCHAR(255) NOT NULL, 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES "users"(id) ON DELETE CASCADE 
);

CREATE INDEX idx_users_phone_number ON "users"(phone_number);
CREATE INDEX idx_contacts_user_id ON "contacts"(user_id);
