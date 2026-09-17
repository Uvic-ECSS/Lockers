CREATE TABLE IF NOT EXISTS locker (
    id varchar(255) NOT NULL,
    PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS registration (
    locker varchar(255) NOT NULL,
    user varchar(255) NOT NULL,
    name varchar(255) NOT NULL,
    expiry datetime NOT NULL,
    expiryEmailSent boolean DEFAULT FALSE,
    PRIMARY KEY (locker)
);

CREATE INDEX IF NOT EXISTS user_registration 
ON registration (user);

CREATE TABLE IF NOT EXISTS history (
    locker varchar(255) NOT NULL,
    user varchar(255) NOT NULL,
    name varchar(255) NOT NULL,
    removed datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS locker_history
ON history (locker, removed);
