CREATE TABLE profiles (
    id SERIAL PRIMARY KEY,
    login VARCHAR(1000) UNIQUE NOT NULL,
    password VARCHAR(1000) NOT NULL,
    first_Name VARCHAR(1000) NOT NULL,
    last_name VARCHAR(1000) NOT NULL,
    middle_name VARCHAR(1000),
    country VARCHAR(100),
    academic_degree VARCHAR(1000),
    VAC VARCHAR(1000),
    appointment VARCHAR(1000)
);

CREATE TABLE publications (
    id SERIAL PRIMARY KEY,
    title VARCHAR(1000) NOT NULL,
    abstract VARCHAR(1000) NOT NULL,
    file_link VARCHAR(1000) NOT NULL,
    created_at DATE NOT NULL,
    updated_at DATE NOT NULL,
    owner_id INTEGER NOT NULL REFERENCES profiles(id)
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
    publications_id INTEGER REFERENCES publications(id) ON DELETE CASCADE,
    tags_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (publications_id, tags_id)
);

CREATE TABLE profile_publications (
    profiles_id INTEGER REFERENCES profiles(id) ON DELETE CASCADE,
    publications_id INTEGER REFERENCES publications(id) ON DELETE CASCADE,
    PRIMARY KEY (profiles_id, publications_id)
);

CREATE TABLE subscribs (
    profiles_id INTEGER REFERENCES Profiles(id),
    subscribers_id INTEGER REFERENCES Profiles(id),
    PRIMARY KEY (profiles_id, subscribers_id)
);
