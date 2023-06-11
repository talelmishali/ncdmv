-- BEGIN;

CREATE TABLE appointment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    location TEXT NOT NULL,
    time DATETIME NOT NULL,
    create_timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE notification_method (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    discord_webhook TEXT,
    create_timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE notification (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    appointment_id INTEGER REFERENCES appointment(id) NOT NULL,
    notification_method_id INTEGER REFERENCES notification_method(id) NOT NULL,
    create_timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- COMMIT;
