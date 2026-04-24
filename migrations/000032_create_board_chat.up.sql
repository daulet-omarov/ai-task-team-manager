CREATE TABLE board_chat_messages (
    id         BIGSERIAL PRIMARY KEY,
    board_id   INT       NOT NULL REFERENCES boards(id)  ON DELETE CASCADE,
    author_id  INT       NOT NULL REFERENCES employees(id),
    reply_to_id BIGINT            REFERENCES board_chat_messages(id) ON DELETE SET NULL,
    text       TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE board_chat_attachments (
    id         BIGSERIAL PRIMARY KEY,
    message_id BIGINT    NOT NULL REFERENCES board_chat_messages(id) ON DELETE CASCADE,
    file_path  TEXT      NOT NULL,
    file_name  TEXT      NOT NULL,
    file_size  INT       NOT NULL,
    mime_type  TEXT      NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE board_polls (
    id         BIGSERIAL PRIMARY KEY,
    message_id BIGINT    NOT NULL REFERENCES board_chat_messages(id) ON DELETE CASCADE,
    question   TEXT      NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE board_poll_options (
    id      BIGSERIAL PRIMARY KEY,
    poll_id BIGINT NOT NULL REFERENCES board_polls(id) ON DELETE CASCADE,
    text    TEXT   NOT NULL
);

CREATE TABLE board_poll_votes (
    id          BIGSERIAL PRIMARY KEY,
    option_id   BIGINT NOT NULL REFERENCES board_poll_options(id) ON DELETE CASCADE,
    employee_id INT    NOT NULL REFERENCES employees(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (option_id, employee_id)
);
