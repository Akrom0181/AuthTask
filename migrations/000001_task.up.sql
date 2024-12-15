CREATE TABLE IF NOT EXISTS "users" (
    id UUID PRIMARY KEY, 
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    phone_number VARCHAR(15) NOT NULL UNIQUE, 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

Create TABLE IF NOT EXISTS "contacts" (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES "users"(id), 
    first_name VARCHAR(55) NOT NULL,
    last_name VARCHAR(55) NOT NULL,
    middle_name VARCHAR(55) NOT NULL,
    phone_number VARCHAR(55) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP
);

CREATE TYPE device_type AS ENUM ('android', 'iOS');

CREATE TABLE IF NOT EXISTS "devices" (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL, 
    name VARCHAR(255) NOT NULL, 
    notification_key VARCHAR(255) UNIQUE NOT NULL,
    type device_type NOT NULL,
    os_version VARCHAR(255) NOT NULL,
    app_version VARCHAR(255) NOT NULL,
    remember_me BOOLEAN DEFAULT false,
    ad_id VARCHAR(255),  
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES "users"(id) ON DELETE CASCADE 
);

CREATE INDEX idx_user_id ON devices(user_id);
CREATE INDEX idx_type ON devices(type); 

CREATE INDEX idx_users_phone_number ON "users"(phone_number);
CREATE INDEX idx_contacts_user_id ON "contacts"(user_id);


