CREATE DATABASE IF NOT EXISTS ether;
USE ether;

CREATE TABLE IF NOT EXISTS conversations (
    ID INTEGER NOT NULL AUTO_INCREMENT,
    Name VARCHAR(255),
    Description VARCHAR(255),
    AvatarURL VARCHAR(255),
    PRIMARY KEY(ID)
);

CREATE TABLE IF NOT EXISTS users_to_conversations (
    UserID INTEGER NOT NULL,
    ConversationID INTEGER NOT NULL,
    Role ENUM('owner', 'admin', 'user'),
    Nickname VARCHAR(255),
    Pending TINYINT(1),
    LastOpened TIMESTAMP,
    FOREIGN KEY (ConversationID) REFERENCES conversations(ID),
    PRIMARY KEY(UserID, ConversationID)
);
