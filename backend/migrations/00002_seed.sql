-- ============================================================================
-- 00002_seed.sql
--
-- Dados iniciais. Idempotente: pode ser reexecutada sem duplicar linhas.
-- Os nomes das comunidades são os da Paróquia São Benedito e Menino Jesus,
-- carregados do schema legado.
-- ============================================================================

INSERT INTO communities (name, description, city) VALUES
    ('Matriz São Benedito', 'Sede da paróquia', 'Francisco Morato'),
    ('Comunidade São Luís Gonzaga', '', 'Francisco Morato'),
    ('Comunidade Santa Rosa de Lima', '', 'Francisco Morato'),
    ('Comunidade Santa Catarina de Sena', '', 'Francisco Morato')
ON CONFLICT DO NOTHING;

INSERT INTO site_settings (id) VALUES (1)
ON CONFLICT (id) DO NOTHING;
