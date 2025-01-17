package postgres

import "solution/internal/domain/entity"

// Migrations is a list of all gorm migrations for the database.
var Migrations = []interface{}{
	&entity.User{},
	&entity.Token{},
	&entity.Business{},
}
