package db

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
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
	log.Info("Users table")
	_, err := db.Exec(ctx, `
	create table users
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

	log.Info("Orders table")
	_, err = db.Exec(ctx, `
	create table orders
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

func CreateUser(db *pgx.Conn, slackId string, name string) (string, error) {
	userId, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	log.Println(userId)
	_, err = db.Exec(context.Background(), "INSERT INTO users(id, name, slack_id) VALUES ($1, $2, $3);", userId, name, slackId)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return userId.String(), nil
}

func FindUserIdBySlackId(db *pgx.Conn, slackId string) (string, error) {
	rows, err := db.Query(context.Background(), "SELECT id FROM users WHERE slack_id = $1;", slackId)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	ids, err := rows.Values()
	if err != nil {
		return "", err
	}
	log.Println(ids)
	if len(ids) == 0 {
		return "-1", err
	}
	var userId uuid.UUID
	userId = ids[0].([16]uint8)
	return userId.String(), nil
}
