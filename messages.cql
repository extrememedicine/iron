CREATE TABLE messages (
id timeuuid,
receiver text,
sender text,
body text,
timestamp timestamp,
PRIMARY KEY (sender, timestamp)
);

CREATE INDEX idIndex ON messages (id);
