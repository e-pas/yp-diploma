package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"yp-diploma/internal/app/config"
	"yp-diploma/internal/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DDL = `
	CREATE TABLE IF NOT EXISTS users (
		id			uuid	NOT NULL CONSTRAINT user_pk PRIMARY KEY,
		name		VARCHAR(20) NOT NULL UNIQUE,
		passwd		CHAR(64)
	);

	CREATE TABLE IF NOT EXISTS session_keys (
		id			CHAR(32) NOT NULL CONSTRAINT keys_pk PRIMARY KEY,
		user_id		uuid 	 NOT NULL REFERENCES users,
		expires		TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS orders (
		id			VARCHAR(20) NOT NULL CONSTRAINT orders_pk PRIMARY KEY,
		user_id		uuid 	 NOT NULL REFERENCES users,
		regdate		TIMESTAMP WITH TIME ZONE NOT NULL,
		accrual		INT
	);
	/* note: orders.accrual values :
			NULL if REGISTERED or PROCESSING, 
			0    if INVALID, 
		   	sum  if PROCESSED
	*/

	CREATE TABLE IF NOT EXISTS withdraws (
		order_id	VARCHAR(20) NOT NULL,
		user_id		uuid 	 NOT NULL REFERENCES users,
		regdate		TIMESTAMP WITH TIME ZONE NOT NULL,
		withdraw	INT NOT NULL		
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
				   FROM withdraws 
			       GROUP BY user_id) AS w 
	        	ON o.user_id = w.user_id);

	/* erase old session keys */				
	DELETE FROM session_keys WHERE expires < NOW();
`

	addUser         = "INSERT INTO users (id, name, passwd) VALUES ($1, $2, $3);"
	getUser         = "SELECT id, name, passwd FROM users WHERE name=$1;"
	addSessKey      = "INSERT INTO session_keys (id, user_id, expires) VALUES ($1, $2, $3);"
	getSessKey      = "SELECT id, user_id, expires FROM session_keys  WHERE id = $1;"
	addOrder        = "INSERT INTO orders (id, user_id, regdate) VALUES ($1, $2, $3);"
	getOrder        = "SELECT id, user_id, regdate, COALESCE(accrual, -1) FROM orders WHERE id = $1;"
	getUserOrders   = "SELECT id, user_id, regdate, COALESCE(accrual, -1) FROM orders WHERE user_id = $1 ORDER BY regdate;"
	getUndoneOrders = "SELECT id, user_id, regdate, COALESCE(accrual, -1) FROM orders WHERE accrual IS NULL;"
	updateAccrual   = "UPDATE orders SET accrual = $2 where id = $1;"
	getBalance      = "SELECT user_id, COALESCE(asum,0), COALESCE(wsum,0), COALESCE(bal,0) FROM balances WHERE user_id = $1;"
	addWithdraw     = "INSERT INTO withdraws (order_id, user_id, regdate, withdraw) VALUES ($1, $2, $3, $4);"
	getWithdraws    = "SELECT order_id, user_id, regdate, withdraw FROM withdraws WHERE user_id = $1 ORDER BY regdate;"
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

	return r.initDDL(ctx)
}

func (r *Repository) initDDL(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, DDL)
	return err
}

func (r *Repository) AddUser(ctx context.Context, user model.User) error {
	var pgerr *pgconn.PgError
	_, err := r.pool.Exec(ctx, addUser, user.ID, user.Name, user.HashedPasswd)
	if err != nil {
		if errors.As(err, &pgerr) && strings.EqualFold(pgerr.ConstraintName, "users_name_key") {
			return config.ErrUserNameBusy
		}
	}
	return err
}

func (r *Repository) GetUserID(ctx context.Context, name string) (model.User, error) {
	var res model.User
	row := r.pool.QueryRow(ctx, getUser, name)
	err := row.Scan(&res.ID, &res.Name, &res.HashedPasswd)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return res, config.ErrNoSuchRecord
		}
		return res, err
	}
	return res, err
}

func (r *Repository) AddSessKey(ctx context.Context, key model.SessKey) error {
	_, err := r.pool.Exec(ctx, addSessKey, key.ID, key.UserID, key.Expires)
	return err
}

func (r *Repository) GetSessKey(ctx context.Context, key string) (model.SessKey, error) {
	var res model.SessKey
	row := r.pool.QueryRow(ctx, getSessKey, key)
	err := row.Scan(&res.ID, &res.UserID, &res.Expires)
	return res, err
}

func (r *Repository) AddOrder(ctx context.Context, order model.Order) error {
	_, err := r.pool.Exec(ctx, addOrder, order.ID, order.UserID, order.GenTime)
	return err
}

func (r *Repository) GetOrder(ctx context.Context, orderID string) (model.Order, error) {
	var res model.Order
	row := r.pool.QueryRow(ctx, getOrder, orderID)
	err := row.Scan(&res.ID, &res.UserID, &res.GenTime, &res.Accrual)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return res, config.ErrNoSuchRecord
		}
		return res, err
	}
	switch {
	case res.Accrual < 0:
		res.Status = model.Created
		res.Accrual = 0
	case res.Accrual == 0:
		res.Status = model.Invalid
	default:
		res.Status = model.Processed
	}
	return res, nil
}

func (r *Repository) GetUserOrders(ctx context.Context, userID string) ([]model.Order, error) {
	return r.getOrderList(ctx, getUserOrders, userID)
}

func (r *Repository) GetUndoneOrders(ctx context.Context) ([]model.Order, error) {
	return r.getOrderList(ctx, getUndoneOrders)
}

func (r *Repository) getOrderList(ctx context.Context, sqlStatement string, param ...any) ([]model.Order, error) {
	res := make([]model.Order, 0)
	rows, err := r.pool.Query(ctx, sqlStatement, param...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		rec := model.Order{}
		err := rows.Scan(&rec.ID, &rec.UserID, &rec.GenTime, &rec.Accrual)
		if err != nil {
			return nil, err
		}
		switch {
		case rec.Accrual < 0:
			rec.Status = model.Created
			rec.Accrual = 0
		case rec.Accrual == 0:
			rec.Status = model.Invalid
		default:
			rec.Status = model.Processed
		}
		res = append(res, rec)
	}
	return res, nil
}

func (r *Repository) UpdateAccruals(ctx context.Context, data []model.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	btch := &pgx.Batch{}

	for _, rec := range data {
		btch.Queue(updateAccrual, rec.ID, rec.Accrual)
	}
	bres := tx.SendBatch(ctx, btch)

	for range data {
		_, qerr := bres.Exec()
		if qerr != nil {
			return qerr
		}
	}
	err = bres.Close()
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetBalance(ctx context.Context, userid string) (model.Balance, error) {
	var res model.Balance
	row := r.pool.QueryRow(ctx, getBalance, userid)
	err := row.Scan(&res.UserID, &res.Accrual, &res.Withdraw, &res.Balance)
	if err != nil {
		if err == pgx.ErrNoRows {
			return res, config.ErrNoSuchRecord
		}
		return res, err
	}
	return res, err
}

func (r *Repository) AddWithdraw(ctx context.Context, w model.Withdraw) error {
	_, err := r.pool.Exec(ctx, addWithdraw, w.OrderID, w.UserID, w.GenTime, w.Withdraw)
	return err
}

func (r *Repository) GetWithdrawList(ctx context.Context, userID string) ([]model.Withdraw, error) {
	res := make([]model.Withdraw, 0)
	rows, err := r.pool.Query(ctx, getWithdraws, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		rec := model.Withdraw{}
		err := rows.Scan(&rec.OrderID, &rec.UserID, &rec.GenTime, &rec.Withdraw)
		if err != nil {
			return nil, err
		}
		res = append(res, rec)
	}
	return res, nil
}
