-- BEGIN;

PRAGMA foreign_keys=ON;

CREATE TABLE appointment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    location TEXT NOT NULL,
    time DATETIME NOT NULL,
    available BOOL NOT NULL DEFAULT false,
    create_timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(location, time)
);

CREATE TABLE notification (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    appointment_id INTEGER REFERENCES appointment(id) ON DELETE CASCADE NOT NULL,
    discord_webhook TEXT,
    available BOOL NOT NULL,
    create_timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- COMMIT;
