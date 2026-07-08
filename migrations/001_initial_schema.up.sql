-- Create films table
CREATE TABLE IF NOT EXISTS films (
    id BIGINT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    year INTEGER,
    runtime INTEGER,
    synopsis TEXT,
    imdb_id VARCHAR(20),
    poster_url VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Seed 10 films
INSERT INTO films (id, title, year, runtime, synopsis, imdb_id, poster_url, created_at) VALUES
(1, 'The Shawshank Redemption', 1994, 142, 'Two imprisoned men bond over a number of years, finding solace and eventual redemption through acts of common decency.', 'tt0111161', 'https://example.com/poster1.jpg', NOW()),
(2, 'The Dark Knight', 2008, 152, 'When the menace known as the Joker wreaks havoc and chaos on the people of Gotham, Batman must accept one of the greatest tests to fight injustice.', 'tt0468569', 'https://example.com/poster2.jpg', NOW()),
(3, 'Inception', 2010, 148, 'A thief who steals corporate secrets through the use of dream-sharing technology is given the inverse task of planting an idea into the mind of a C.E.O.', 'tt1375666', 'https://example.com/poster3.jpg', NOW()),
(4, 'The Matrix', 1999, 136, 'A computer hacker learns from mysterious rebels about the true nature of his reality and his role in the war against its controllers.', 'tt0133093', 'https://example.com/poster4.jpg', NOW()),
(5, 'Pulp Fiction', 1994, 154, 'The lives of two mob hitmen, a boxer, a gangsters wife, and a pair of diner bandits intertwine in four tales of violence and redemption.', 'tt0110912', 'https://example.com/poster5.jpg', NOW()),
(6, 'Forrest Gump', 1994, 142, 'The presidencies of Kennedy and Johnson, the Vietnam War, the Watergate scandal and other historical events unfold from the perspective of an Alabama man with an IQ of 75.', 'tt0109830', 'https://example.com/poster6.jpg', NOW()),
(7, 'Interstellar', 2014, 169, 'A team of explorers travel through a wormhole in space in an attempt to ensure humanity''s survival.', 'tt0816692', 'https://example.com/poster7.jpg', NOW()),
(8, 'The Godfather', 1972, 175, 'The aging patriarch of an organized crime dynasty transfers control of his clandestine empire to his reluctant youngest son.', 'tt0068646', 'https://example.com/poster8.jpg', NOW()),
(9, 'Gladiator', 2000, 155, 'A former Roman General sets out to exact vengeance against the corrupt emperor who murdered his family and sent him into slavery.', 'tt0172495', 'https://example.com/poster9.jpg', NOW()),
(10, 'The Lord of the Rings: The Fellowship of the Ring', 2001, 178, 'A meek Hobbit from the Shire and eight companions embark on a journey to Mount Doom to destroy the One Ring and defeat its maker, the Dark Lord Sauron.', 'tt0120737', 'https://example.com/poster10.jpg', NOW())
ON CONFLICT DO NOTHING;
