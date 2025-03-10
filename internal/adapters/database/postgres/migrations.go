package postgres

import "prod/internal/domain/entity"

// Migrations is a list of all gorm migrations for the database.
var Migrations = []interface{}{
	&entity.User{},
	&entity.Business{},
	&entity.Promo{},
	&entity.PromoUnique{},
	&entity.Category{},
	&entity.Likes{},
	&entity.Comment{},
	&entity.Activation{},
}
