SET FOREIGN_KEY_CHECKS=0;
DROP TABLE IF EXISTS meme;
DROP TABLE IF EXISTS meme_keyword;
DROP TABLE IF EXISTS system_metadata;
DROP TABLE IF EXISTS temp_files;
SET FOREIGN_KEY_CHECKS=1;

CREATE TABLE IF NOT EXISTS meme (
	id INT NOT NULL AUTO_INCREMENT,
	source ENUM('NONE', 'IMGUR'),
	url VARCHAR(255) DEFAULT NULL,
	top_text VARCHAR(255) NOT NULL DEFAULT '',
	bottom_text VARCHAR(255) NOT NULL DEFAULT '',
	net_ups INT,
	views INT,
	num_keywords INT,
	meme_name VARCHAR(100) DEFAULT NULL,
	imgur_bg_image VARCHAR(100) DEFAULT NULL,
	PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS meme_keyword (
	meme_id INT NOT NULL,
	keyword VARCHAR(50) NOT NULL,
	PRIMARY KEY (meme_id, keyword),
	INDEX keyword_idx (keyword),
	FOREIGN KEY (meme_id) REFERENCES meme(id)
);

CREATE TABLE IF NOT EXISTS system_metadata (
	message_id VARCHAR(128) NOT NULL,
	messenger_id VARCHAR(128) NOT NULL DEFAULT '',
	meme_id INT NOT NULL,
	is_upvote BOOL DEFAULT NULL,
	PRIMARY KEY( message_id ),
	INDEX messenger_id_idx (messenger_id),
	FOREIGN KEY (meme_id) REFERENCES meme(id)
);

CREATE TABLE IF NOT EXISTS temp_files (
  id INT NOT NULL AUTO_INCREMENT,
  message_id VARCHAR(128) NOT NULL,
  file_name VARCHAR(50) NOT NULL,
  time_created TIMESTAMP NOT NULL,
  PRIMARY KEY ( id ),
  INDEX message_id_idx (message_id)
);
