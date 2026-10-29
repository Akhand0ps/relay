package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/akhand0ps/relay/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
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

var ErrDuplicateEvent = errors.New("event already exists with this idempotency key")

const UniqueViolationCode = "23505"

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

	if err != nil{

		
		var pgErr *pgconn.PgError
        //unwrapp and check the err type
		if errors.As(err,&pgErr){
			
			if pgErr.Code == UniqueViolationCode && pgErr.ConstraintName == "unique_tenant_idempotency"{

				return ErrDuplicateEvent
			}
		}


		return fmt.Errorf("failed to save event: %w",err)
	}

	return nil
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

func (p *PostgresStore) SaveEndpoint(ep models.Endpoint) error {


	ctx,cancel := context.WithTimeout(context.Background(),5 * time.Second)
	defer cancel()

	query := `
	 
	INSERT INTO endpoints(id,url,event_types,secret,status,created_at)
	VALUES ($1,$2,$3,$4,$5,$6)
	`

	_,err := p.pool.Exec(ctx,query,
		ep.ID,
		ep.URL,
		ep.EventTypes,
		ep.Secret,
		ep.Status,
		ep.CreatedAt,
	)

	if err != nil{
		return fmt.Errorf("failed to save endpoint: %w",err)
	}

	return nil
}

func(p *PostgresStore) GetEndpoint(id string) (*models.Endpoint,error){
	ctx,cancel := context.WithTimeout(context.Background(),5 * time.Second)
	defer cancel()


	query:= `
	
	SELECT id,url,event_types,secret,status,created_at
	FROM endpoints
	WHERE id = $1
	`

	var ep models.Endpoint

	err:= p.pool.QueryRow(ctx,query,id).Scan(
		&ep.ID,
		&ep.URL,
		&ep.EventTypes,
		&ep.Secret,
		&ep.Status,
		&ep.CreatedAt,
	)


	if err != nil{

		return nil,err
	}

	return &ep,nil

}


// list endpoints

func(p *PostgresStore) ListEndpoint()([]models.Endpoint,error){

	ctx,cancel := context.WithTimeout(context.Background(),5 * time.Second)

	defer cancel()



	query := `
	
	SELECT id, url, event_types, secret, status, created_at FROM endpoints
	`

	rows,err := p.pool.Query(ctx,query)

	if err != nil {
		return nil,err
	}

	defer rows.Close()

	var endpoints []models.Endpoint

	for rows.Next() {

		var ep models.Endpoint

		err := rows.Scan(
			&ep.ID,
			&ep.URL,
			&ep.EventTypes,
			&ep.Secret,
			&ep.Status,
			&ep.CreatedAt,
		)

		if err != nil{
			return  nil,err
		}

		endpoints = append(endpoints,ep)
	}
	return endpoints,nil
}


func (p *PostgresStore) FindEndPointForEvent(eventType string)([]models.Endpoint,error){

	ctx, cancel := context.WithTimeout(context.Background(),5 * time.Second)

	defer cancel()

	query:= `

	SELECT id, url, event_types, secret, status, created_at 
	FROM endpoints 
	WHERE status = 'active' 
	AND $1 = ANY(event_types);	
	`

	rows, err := p.pool.Query(ctx,query,eventType)

	if err != nil {
		return nil,err
	}

	defer rows.Close()

	var endpoints []models.Endpoint

	for rows.Next() {

		var ep models.Endpoint

		err := rows.Scan(
			&ep.ID,
			&ep.URL,
			&ep.EventTypes,
			&ep.Secret,
			&ep.Status,
			&ep.CreatedAt,
		)

		if err != nil{
			return  nil,err
		}

		endpoints = append(endpoints,ep)
	}
	return endpoints,nil

}

func (p *PostgresStore) GetByIdempotencyKey(key string)(*models.Event,bool){

	ctx,cancel := context.WithTimeout(context.Background(),5 * time.Second)
	defer cancel()

	query := `
      
	SELECT id,tenant_id,type,payload,idempotency_key,status,created_at
	FROM events
	WHERE tenant_id = $1 AND idempotency_key = $2
	`

	var event models.Event

	err:= p.pool.QueryRow(ctx,query,"default_tenant",key).Scan(
		&event.ID,
		&event.TenantID,
		&event.Type,
		&event.Payload,
		&event.IdempotencyKey,
		&event.Status,
		&event.CreatedAt,
	)

	if err != nil{
		return nil,false
	}
	return &event,true
}


func (p *PostgresStore) SaveDelivery(d models.Delivery) error{

	ctx,cancel := context.WithTimeout(context.Background(),5 * time.Second)
	defer cancel()

	query := `

	INSERT INTO deliveries (id, event_id, endpoint_id, status, attempt_count, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)

	`

	_,err:= p.pool.Exec(ctx,query,
		d.ID,
		d.EventId,
		d.EndpointId,
		d.Status,
		d.AttemptCount,
		d.UpdatedAt,
	)

	return err

}

func (p *PostgresStore) RecordDeliveryAttempt(attempt models.DeliveryAttempt) error{
	ctx,cancel:= context.WithTimeout(context.Background(),5 * time.Second)
	defer cancel()

	query := `
     INSERT INTO delivery_attempts (id, delivery_id, status_code, error_message, attempted_at)
     VALUES ($1, $2, $3, $4, $5)
`

	_, err := p.pool.Exec(ctx, query,
		attempt.ID,
		attempt.DeliveryId,
		attempt.StatusCode,
		attempt.ErrorMessage,
		attempt.AttemptedAt,
	)
	return err
}	
func (p *PostgresStore) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}
