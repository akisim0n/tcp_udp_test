package user

import (
	"context"
	"fmt"
	"strings"
	"tcp_upd_test/tcp/http/models"
	"tcp_upd_test/tcp/http/repository"
	"tcp_upd_test/tcp/http/repository/builders"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = "users"

type db interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type repo struct {
	DB db
}

func NewRepository(db *pgxpool.Pool) repository.UserRepository {
	return &repo{DB: db}
}

func (r *repo) Get(ctx context.Context, id int64) (*models.User, error) {

	user := &models.User{}

	bBuilder := builders.NewBatchBuilder()
	getQ := bBuilder.AddQueryRow(&user.ID, &user.Name, &user.Surname, &user.Email, &user.Age, &user.CreatedAt, &user.UpdatedAt)
	getQ(fmt.Sprintf("SELECT id, name, surname, email, age, created_at, updated_at FROM %s WHERE id = $1", tableName), id)

	batchErr := r.WithTxWithBatch(ctx, bBuilder)
	if batchErr != nil {
		return nil, fmt.Errorf("get: %w", batchErr)
	}

	return user, nil
}
func (r *repo) GetAll(ctx context.Context) ([]*models.User, error) {

	rowsData, execErr := r.DB.Query(ctx, "SELECT id, name, surname, age, updated_at, created_at, email FROM users")
	if execErr != nil {
		return nil, fmt.Errorf("getAll: %w", execErr)
	}

	var users []*models.User

	for rowsData.Next() {
		user := models.User{}
		if scanErr := rowsData.Scan(&user.ID, &user.Name, &user.Surname, &user.Age, &user.UpdatedAt, &user.CreatedAt, &user.Email); scanErr != nil {
			return nil, scanErr
		}
		users = append(users, &user)
	}
	defer rowsData.Close()

	return users, nil
}
func (r *repo) Update(ctx context.Context, user *models.User) error {
	userId := user.ID
	var queryArgs []string
	var queryValue []interface{}
	argsPos := 1

	if user.Name != "" {
		queryArgs = append(queryArgs, fmt.Sprintf("name = $%d", argsPos))
		queryValue = append(queryValue, user.Name)
		argsPos++
	}

	if user.Age != nil {
		queryArgs = append(queryArgs, fmt.Sprintf("age = $%d", argsPos))
		queryValue = append(queryValue, *user.Age)
		argsPos++
	}

	if user.Surname != nil {
		queryArgs = append(queryArgs, fmt.Sprintf("surname = $%d", argsPos))
		queryValue = append(queryValue, *user.Surname)
		argsPos++
	}

	if user.Email != "" {
		queryArgs = append(queryArgs, fmt.Sprintf("email = $%d", argsPos))
		queryValue = append(queryValue, user.Email)
		argsPos++
	}

	if user.UpdatedAt.IsZero() && len(queryArgs) == 0 {
		return nil
	}

	if !user.UpdatedAt.IsZero() {
		queryArgs = append(queryArgs, fmt.Sprintf("updated_at = $%d", argsPos))
		queryValue = append(queryValue, user.UpdatedAt)
		argsPos++
	}

	queryText := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", tableName, strings.Join(queryArgs, ", "), argsPos)

	queryValue = append(queryValue, userId)

	bBuilder := builders.NewBatchBuilder()
	updateQ := bBuilder.AddExec()
	updateQ(queryText, queryValue...)

	return r.WithTxWithBatch(ctx, bBuilder)
}
func (r *repo) Create(ctx context.Context, user *models.User) (int64, error) {
	var newUserId int64
	builder := builders.NewInsertBuilder(tableName)
	bBuilder := builders.NewBatchBuilder()

	builder.Set("name", user.Name)
	builder.Set("email", user.Email)
	builder.Set("password", user.Password)
	builder.Set("created_at", user.CreatedAt)
	builder.Set("updated_at", user.UpdatedAt)

	if user.Surname != nil {
		builder.Set("surname", *user.Surname)
	}
	if user.Age != nil {
		builder.Set("age", *user.Age)
	}

	builder.SetReturning("id")

	queryText, args := builder.Build()

	insertQ := bBuilder.AddQueryRow(&newUserId)
	insertQ(queryText, args...)

	execErr := r.WithTxWithBatch(ctx, bBuilder)
	if execErr != nil {
		return 0, fmt.Errorf("error executing insert query: %s", execErr)
	}

	return newUserId, nil
}
func (r *repo) Delete(ctx context.Context, id int64) error {

	queryText := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName)

	execErr := r.WithTx(ctx, func(tx pgx.Tx) error {
		_, txErr := tx.Exec(ctx, queryText, id)
		return txErr
	})
	if execErr != nil {
		return fmt.Errorf("error while deleting user: %s", execErr.Error())
	}
	return nil
}

func (r *repo) WithTxWithBatch(ctx context.Context, bb *builders.BatchBuilder) error {
	tx, txBeginErr := r.DB.Begin(ctx)
	if txBeginErr != nil {
		return fmt.Errorf("error starting transaction: %s", txBeginErr)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	batchErr := bb.Run(ctx, tx)
	if batchErr != nil {
		return fmt.Errorf("error batching transaction: %s", batchErr)
	}

	return tx.Commit(ctx)
}

func (r *repo) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, txBeginErr := r.DB.Begin(ctx)
	if txBeginErr != nil {
		return fmt.Errorf("error starting transaction: %s", txBeginErr)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if execErr := fn(tx); execErr != nil {
		return execErr
	}

	return tx.Commit(ctx)
}
