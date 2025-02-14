CREATE TABLE profiles (
    id SERIAL PRIMARY KEY,
    login VARCHAR(1000) UNIQUE NOT NULL,
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

CREATE TABLE publications (
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

CREATE TABLE publication_tags (
    publications_id INTEGER REFERENCES publications(id),
    tags_id INTEGER REFERENCES tags(id),
    PRIMARY KEY (publications_id, tags_id)
);

CREATE TABLE profile_publications (
    profiles_id INTEGER REFERENCES profiles(id),
    publications_id INTEGER REFERENCES publications(id),
    PRIMARY KEY (profiles_id, publications_id)
);

CREATE TABLE subscribs (
    profile_id INTEGER REFERENCES Profiles(id),
    subscriber_id INTEGER REFERENCES Profiles(id),
    PRIMARY KEY (profile_id, subscriber_id)
);
