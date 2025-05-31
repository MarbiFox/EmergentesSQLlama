-- Table: Usuario
CREATE DATABASE IF NOT EXISTS datasql;
USE datasql;

-- Table: ModeloLLM
CREATE TABLE ModeloLLM (
    idModelo INT AUTO_INCREMENT PRIMARY KEY,
    NombreModelo VARCHAR(100) NOT NULL,
    DatosFineTunning TEXT
);

CREATE TABLE Usuario (
    idUsuario INT AUTO_INCREMENT PRIMARY KEY,
    RUT VARCHAR(20) NOT NULL UNIQUE,
    Nombre VARCHAR(100) NOT NULL,
    idModelo INT NOT NULL,
    FOREIGN KEY (idModelo) REFERENCES ModeloLLM(idModelo) ON DELETE CASCADE
);

-- Table: Consulta (Prompt)
CREATE TABLE Consulta (
    idPrompt INT AUTO_INCREMENT PRIMARY KEY,
    Texto TEXT NOT NULL,
    idUsuario INT NOT NULL,
    FOREIGN KEY (idUsuario) REFERENCES Usuario(idUsuario) ON DELETE CASCADE
);

-- Table: Respuesta
CREATE TABLE Respuesta (
    idRespuesta INT AUTO_INCREMENT PRIMARY KEY,
    Texto TEXT NOT NULL,
    idPrompt INT NOT NULL,
    idModelo INT NOT NULL,
    FOREIGN KEY (idPrompt) REFERENCES Consulta(idPrompt) ON DELETE CASCADE,
    FOREIGN KEY (idModelo) REFERENCES ModeloLLM(idModelo) ON DELETE CASCADE
);

INSERT INTO ModeloLLM (NombreModelo, DatosFineTunning) VALUES ('Llama3.2', 'Modelo de Prueba');
INSERT INTO ModeloLLM (NombreModelo, DatosFineTunning) VALUES ('SQLlama', 'Modelo con Fine-Tuning');
INSERT INTO Usuario (RUT, Nombre, idModelo) VALUES ('12345678-9', 'Juan Pérez', 1);
INSERT INTO Usuario (RUT, Nombre, idModelo) VALUES ('21151511-2', 'Vicente Mercado', 2);
INSERT INTO Consulta (Texto, idUsuario) VALUES ('¿Cuál es la capital de Francia?', 1);
INSERT INTO Respuesta (Texto, idPrompt, idModelo) VALUES ('La capital de Francia es París.', 1, 1);

/*
INSERT INTO Usuario (RUT, Nombre) VALUES ('12345678-9', 'Juan Pérez');
SELECT * FROM Usuario WHERE idUsuario = 1;
UPDATE Usuario SET Nombre = 'Juan P. Gonzalez' WHERE idUsuario = 1;
DELETE FROM Usuario WHERE idUsuario = 1;

INSERT INTO ModeloLLM (NombreModelo, DatosFineTunning) 
VALUES ('llama3-8b', 'Custom fine-tune for legal prompts');
SELECT * FROM ModeloLLM WHERE idModelo = 1;
UPDATE ModeloLLM 
SET NombreModelo = 'llama3-8b-v2', DatosFineTunning = 'Updated FT data' 
WHERE idModelo = 1;
DELETE FROM ModeloLLM WHERE idModelo = 1;

INSERT INTO Consulta (Texto, idUsuario) 
VALUES ('¿Cuál es la capital de Francia?', 1);
SELECT * FROM Consulta WHERE idPrompt = 1;
UPDATE Consulta 
SET Texto = '¿Cuál es la capital de Alemania?' 
WHERE idPrompt = 1;
DELETE FROM Consulta WHERE idPrompt = 1;

INSERT INTO Respuesta (Texto, idPrompt, idModelo) 
VALUES ('La capital de Francia es París.', 1, 1);
SELECT * FROM Respuesta WHERE idRespuesta = 1;
UPDATE Respuesta 
SET Texto = 'La capital de Alemania es Berlín.' 
WHERE idRespuesta = 1;
DELETE FROM Respuesta WHERE idRespuesta = 1;


SELECT u.Nombre AS Usuario, c.Texto AS Pregunta, r.Texto AS Respuesta, m.NombreModelo AS Modelo
FROM Respuesta r
JOIN Consulta c ON r.idPrompt = c.idPrompt
JOIN Usuario u ON c.idUsuario = u.idUsuario
JOIN ModeloLLM m ON r.idModelo = m.idModelo;
*/