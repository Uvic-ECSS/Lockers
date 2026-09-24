CREATE TABLE IF NOT EXISTS lockers (
    locker_id varchar(255) NOT NULL,
    PRIMARY KEY(locker_id)
);

CREATE TABLE IF NOT EXISTS locker_registrations (
    locker_id varchar(255) NOT NULL,
    user_email varchar(255) NOT NULL,
    user_name varchar(255) NOT NULL,
    expiry_date datetime NOT NULL,
    expiry_email_sent boolean DEFAULT FALSE,
    PRIMARY KEY (locker_id)
);

CREATE INDEX IF NOT EXISTS idx_locker_registrations_user_email
ON locker_registrations(user_email);

CREATE TABLE IF NOT EXISTS locker_removals (
    removal_id INTEGER PRIMARY KEY,
    locker_id varchar(255) NOT NULL,
    user_email varchar(255) NOT NULL,
    user_name varchar(255) NOT NULL,
    removed_date datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_locker_removals_locker_id
ON locker_removals(locker_id);
