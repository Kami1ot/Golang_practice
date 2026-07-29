-- Прогресс и пользователи: только сырые факты.
-- Замки, доступность, XP и условия ачивок всегда вычисляются из фактов + текущего контента.
-- Скрипт идемпотентен: выполняется целиком на каждом старте сервера.

CREATE TABLE IF NOT EXISTS users (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username      text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    role          text NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
    token      text PRIMARY KEY,
    user_id    bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS sessions_expires_idx ON sessions (expires_at);

-- Свежие установки сразу получают финальную форму (no-op, если таблицы уже есть).
CREATE TABLE IF NOT EXISTS level_progress (
    user_id       bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id     text NOT NULL,
    level_id      text NOT NULL,
    quiz_passed   boolean NOT NULL DEFAULT false,
    quiz_attempts integer NOT NULL DEFAULT 0,
    updated_at    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, course_id, level_id)
);

CREATE TABLE IF NOT EXISTS task_progress (
    user_id      bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id    text NOT NULL,
    level_id     text NOT NULL,
    task_id      text NOT NULL,
    completed    boolean NOT NULL DEFAULT false,
    attempts     integer NOT NULL DEFAULT 0,
    draft        text NOT NULL DEFAULT '',
    completed_at timestamptz,
    updated_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, course_id, level_id, task_id)
);

-- Апгрейд БД версии v1 (до пользователей): добавляем НУЛЛИМЫЙ user_id.
-- PK невозможен, пока строки-сироты держат NULL, поэтому вместо него UNIQUE INDEX —
-- ON CONFLICT работает по нему точно так же. Сирот усыновляет первый
-- зарегистрированный пользователь (транзакция CreateUser).
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name = 'level_progress' AND column_name = 'user_id') THEN
    ALTER TABLE level_progress ADD COLUMN user_id bigint REFERENCES users(id) ON DELETE CASCADE;
    ALTER TABLE level_progress DROP CONSTRAINT level_progress_pkey;
    CREATE UNIQUE INDEX level_progress_user_uq ON level_progress (user_id, course_id, level_id);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name = 'task_progress' AND column_name = 'user_id') THEN
    ALTER TABLE task_progress ADD COLUMN user_id bigint REFERENCES users(id) ON DELETE CASCADE;
    ALTER TABLE task_progress DROP CONSTRAINT task_progress_pkey;
    CREATE UNIQUE INDEX task_progress_user_uq ON task_progress (user_id, course_id, level_id, task_id);
  END IF;
END $$;

-- История чата с наставником (per user + уровень).
CREATE TABLE IF NOT EXISTS chat_messages (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id  text NOT NULL,
    level_id   text NOT NULL,
    role       text NOT NULL CHECK (role IN ('user', 'assistant')),
    content    text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS chat_messages_topic_idx ON chat_messages (user_id, course_id, level_id, id);

-- Ачивки: хранится только факт выдачи; условия — в коде.
CREATE TABLE IF NOT EXISTS user_achievements (
    user_id        bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id text NOT NULL,
    granted_at     timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, achievement_id)
);

-- Заметки пользователя к уровням (Markdown, одна на уровень).
CREATE TABLE IF NOT EXISTS notes (
    user_id    bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id  text NOT NULL,
    level_id   text NOT NULL,
    body       text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, course_id, level_id)
);
