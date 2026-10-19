package store

import (
	"context"
	"fmt"
	"time"

	"github.com/akhand0ps/relay/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)


type PostgresStore struct{

	//A Pool lets dozens of Goroutines talk to Postgres 
	//simultaneously without opening and closing TCP sockets!
	pool *pgxpool.Pool
}
//this is for initial db setup
func NewPostgresStore(connStr string) (*PostgresStore,error){

	pool,err := pgxpool.New(context.Background(),connStr)
	
	if err != nil{
		return nil,fmt.Errorf("unable to connect to database: %w",err)
	}

	return &PostgresStore{pool:pool},nil
}

func (p * PostgresStore) Save(event models.Event) error{

	ctx,cancel:= context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()



	query := `

	INSERT INTO events (id ,tenant_id, type,payload,idempotency_key, status,created_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7)
	`

	_,err := p.pool.Exec(ctx,query,
		event.ID,
		"default_tenant",
		event.Type,
		event.Payload,
		event.IdempotencyKey,
		event.Status,
		event.CreatedAt,
	)

	return err
}


func (p * PostgresStore) Get(id string) (*models.Event,error){

	ctx,cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()


	query := `
	
	SELECT id,tenant_id,type,payload,idempotency_key,status,created_at
	FROM events
	WHERE id = $1
	`

	var event models.Event

	err := p.pool.QueryRow(ctx,query,id).Scan(

		&event.ID,
		&event.TenantID,
		&event.Type,
		&event.Payload,
		&event.IdempotencyKey,
		&event.Status,
		&event.CreatedAt,
	)

	if err != nil{
		return nil,err
	}

	return &event,nil
}


