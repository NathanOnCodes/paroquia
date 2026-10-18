package database

import "embed"

// MigrationsSQL contém as migrations versionadas do banco.
//
// Elas são a fonte de verdade do schema: o Ent gera apenas as consultas
// tipadas. Manter o DDL aqui, e não derivado dos schemas, mantém o banco
// revisável em revisão de código.
//
//go:embed migrations/*.sql
var MigrationsSQL embed.FS
