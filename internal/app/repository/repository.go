package repository

import (
	"context"
	"fmt"

	"yp-diploma/internal/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DDL = `
	CREATE TABLE IF NOT EXISTS users (
		id		uuid	NOT NULL CONSTRAINT user_pk PRIMARY KEY,
		name		VARCHAR(20) NOT NULL UNIQUE,
		passwd		CHAR(64)
	);

	CREATE TABLE IF NOT EXISTS session_keys (
		id		CHAR(32) NOT NULL CONSTRAINT keys_pk PRIMARY KEY,
		user_id		uuid 	 NOT NULL REFERENCES users,
		expires		TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS orders (
		id		VARCHAR(20) NOT NULL CONSTRAINT orders_pk PRIMARY KEY,
		user_id		uuid 	 NOT NULL REFERENCES users,
		regdate		TIMESTAMP NOT NULL,
		accrual		int
	);
	/* note: orders.accrual values :
			NULL if REGISTERED or PROCESSING, 
			0    if INVALID, 
		   	sum  if PROCESSED
	*/

	CREATE TABLE IF NOT EXISTS withdraws (
		order_id	VARCHAR(20) NOT NULL REFERENCES orders,
		regdate		TIMESTAMP NOT NULL,
		withdraw	int NOT NULL		
	);

	CREATE OR REPLACE VIEW balances AS 
		 (SELECT o.user_id, 
			 o.asum, 
			 w.wsum, 
			 (COALESCE(o.asum, 0) - COALESCE(w.wsum, 0)) bal 
		    FROM (SELECT user_id, 
				 sum(accrual) asum
			    FROM orders
			    GROUP BY user_id) AS o
		    LEFT JOIN 
			 (SELECT user_id, 
				 sum(withdraw) wsum 
			    FROM orders, withdraws 
			    WHERE orders.id = withdraws.order_id
			    GROUP BY user_id) AS w 
	        ON o.user_id = w.user_id);

	/* erase old session keys */				
	DELETE FROM session_keys WHERE expires < NOW();
`

	addUser    = "INSERT INTO users (id, name, passwd) VALUES ($1, $2, $3);"
	getUser    = "SELECT id, name, passwd FROM users WHERE name=$1;"
	addSessKey = "INSERT INTO session_keys (id, user_id, expires) VALUES ($1, $2, $3);"
	getSessKey = "SELECT id, user_id, expires FROM session_keys  WHERE id = $1;"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) Init(ctx context.Context, conn string) error {
	cfg, err := pgxpool.ParseConfig(conn)
	if err != nil {
		return fmt.Errorf("parsing conn string: %s failed. error: %w", conn, err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("unable connect to database. error: %w", err)
	}
	r.pool = pool

	return r.InitDDL(ctx)
}

func (r *Repository) InitDDL(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, DDL)
	return err
}

func (r *Repository) AddUser(ctx context.Context, user model.User) error {
	_, err := r.pool.Exec(ctx, addUser, user.ID, user.Name, user.HashedPasswd)
	return err
}

func (r *Repository) GetUserID(ctx context.Context, name string) (model.User, error) {
	var res model.User
	row := r.pool.QueryRow(ctx, getUser, name)
	err := row.Scan(&res.ID, &res.Name, &res.HashedPasswd)
	return res, err
}

func (r *Repository) AddSessKey(ctx context.Context, key model.SessKey) error {
	_, err := r.pool.Exec(ctx, addSessKey, key.ID, key.User_ID, key.Expires)
	return err
}

func (r *Repository) GetSessKey(ctx context.Context, key string) (model.SessKey, error) {
	var res model.SessKey
	row := r.pool.QueryRow(ctx, getSessKey, key)
	err := row.Scan(&res.ID, &res.User_ID, &res.Expires)
	return res, err
}
