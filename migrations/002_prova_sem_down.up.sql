-- PR de prova do gate de migrations (task 0004): migration SEM par .down.sql.
-- O job condicional deve rodar (path-filter positivo) e falhar no down -all.
CREATE TABLE prova_descartavel (id INT PRIMARY KEY);
