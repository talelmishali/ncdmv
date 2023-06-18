-- BEGIN;

CREATE TABLE appointment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    location TEXT NOT NULL,
    time DATETIME NOT NULL,
    create_timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(location, time)
);

CREATE TABLE notification (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    appointment_id INTEGER REFERENCES appointment(id) NOT NULL,
    discord_webhook TEXT,
    create_timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(appointment_id, discord_webhook)
);

-- COMMIT;
