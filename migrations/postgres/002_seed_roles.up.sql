INSERT INTO roles (name, is_default)
VALUES ('admin' , false),
       ('moderator' , false),
       ('user', true)
ON CONFLICT (name) DO NOTHING;