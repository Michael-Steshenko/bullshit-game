CREATE TABLE questions (
    id          SERIAL PRIMARY KEY,
    lang        TEXT NOT NULL DEFAULT 'en',
    question    TEXT NOT NULL,
    real_answer TEXT NOT NULL,
    fake_answers TEXT[] NOT NULL,
    citation    TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_questions_lang ON questions(lang);
