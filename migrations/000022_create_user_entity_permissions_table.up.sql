CREATE TABLE user_entity_permissions
(
    id   SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users (id),
    entity_permission_id INTEGER REFERENCES entity_permissions (id)
);