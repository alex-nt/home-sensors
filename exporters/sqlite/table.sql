CREATE TABLE IF NOT EXISTS measurement (
    id TEXT PRIMARY KEY,
    description TEXT,
    unit TEXT
);

CREATE TABLE IF NOT EXISTS measurement_recording (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    value REAL,
    measure_id TEXT,
    FOREIGN KEY(measure_id) REFERENCES measurement(id)
)

CREATE TABLE IF NOT EXISTS measurement_meta_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT,
    value TEXT,
    recording_id INTEGER,
    FOREIGN KEY(recording_id) REFERENCES measurement_recording(id)
);