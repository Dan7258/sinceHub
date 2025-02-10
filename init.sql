CREATE TABLE Profiles (
    id SERIAL PRIMARY KEY,
    login VARCHAR(1000) NOT NULL,
    password VARCHAR(1000) NOT NULL,
    first_Name VARCHAR(1000) NOT NULL,
    last_name VARCHAR(1000) NOT NULL,
    middle_name VARCHAR(1000),
    Country VARCHAR(100),
    academic_degree VARCHAR(1000),
    VAC VARCHAR(1000),
    appointment VARCHAR(1000),
    subscribers INTEGER,
    my_subscribes INTEGER
);

CREATE TABLE Publications (
    id SERIAL PRIMARY KEY,
    title VARCHAR(1000) NOT NULL,
    abstract VARCHAR(1000),
    content TEXT NOT NULL,
    created_at DATE NOT NULL,
    updated_at DATE NOT NULL
);

CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(1000) UNIQUE NOT NULL
);

CREATE TABLE friend_requests (
    id SERIAL PRIMARY KEY,
    requester INTEGER NOT NULL,
    receiver INTEGER NOT NULL,
    status VARCHAR(100) NOT NULL,
    created_at DATE NOT NULL
);

CREATE TABLE Pubication_Tags (
    publication_id INTEGER REFERENCES Publications(id),
    tag_id INTEGER REFERENCES tags(id),
    PRIMARY KEY (publication_id, tag_id)
);

CREATE TABLE Profile_Publication (
    profile_id INTEGER REFERENCES Profiles(id),
    publication_id INTEGER REFERENCES Publications(id),
    PRIMARY KEY (profile_id, publication_id)
);

CREATE TABLE subscribs (
    profile_id INTEGER REFERENCES Profiles(id),
    subscriber_id INTEGER REFERENCES Profiles(id),
    PRIMARY KEY (profile_id, subscriber_id)
);
