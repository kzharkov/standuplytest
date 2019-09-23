package db

import (
	"context"
	"github.com/jackc/pgx/v4"
	"time"
)

func Conn(config string) (*pgx.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return pgx.Connect(ctx, config)
}

func CreateModels(db *pgx.Conn) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, err := db.Exec(ctx, `create table users
	(
		id uuid not null
			constraint users_pk
				primary key,
		name varchar not null,
		slack_id varchar not null
	);`)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, `create table orders
	(
		user_id uuid not null
			constraint orders_users_id_fk
				references users
					on update cascade,
		number serial not null
			constraint orders_pk
				primary key,
		pizza varchar not null,
		size int not null,
		address varchar not null
	);`)
	if err != nil {
		return err
	}

	return nil
}
