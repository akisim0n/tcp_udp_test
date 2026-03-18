package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"tcp_upd_test/tcp/http/models"
	"tcp_upd_test/tcp/http/repository"
	"tcp_upd_test/tcp/http/repository/builders"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const tableName = "users"

type repo struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) repository.UserRepository {
	return &repo{DB: db}
}

func (r *repo) Get(ctx context.Context, id int64) (*models.User, error) {

	user := &models.User{}

	bBuilder := builders.NewBatchBuilder()
	getQ := bBuilder.AddQuery(&user.ID, &user.Data.Name, &user.Data.Surname, &user.Data.Email, &user.Data.Age, &user.CreatedAt, &user.UpdatedAt)
	getQ(fmt.Sprintf("SELECT id, name, surname, email, age, created_at, updated_at FROM %s WHERE id = $1", tableName), id)

	batchErr := r.WithTxWithBatch(ctx, bBuilder)
	if batchErr != nil {
		return nil, fmt.Errorf("get: %w", batchErr)
	}

	return user, nil
}
func (r *repo) GetAll(ctx context.Context) ([]*models.User, error) {

	var rowsData pgx.Rows

	bBuilder := builders.NewBatchBuilder()
	getQ := bBuilder.AddQuery(&rowsData)
	getQ("SELECT * FROM users")

	batchErr := r.WithTxWithBatch(ctx, bBuilder)
	if batchErr != nil {
		return nil, fmt.Errorf("getAll: %w", batchErr)
	}

	var users []*models.User

	for rowsData.Next() {
		var user models.User
		if scanErr := rowsData.Scan(&user); scanErr != nil {
			return nil, scanErr
		}
		users = append(users, &user)
	}

	return users, nil
}
func (r *repo) Update(ctx context.Context, user *models.User) error {

	userId := user.ID
	var queryArgs []string
	var queryValue []interface{}
	argsPos := 1

	if user.Data.Name != "" {
		queryArgs = append(queryArgs, fmt.Sprintf("name = &%d", argsPos))
		queryValue = append(queryValue, user.Data.Name)
		argsPos++
	}

	if user.Data.Age != nil {
		queryArgs = append(queryArgs, fmt.Sprintf("age = &%d", argsPos))
		queryValue = append(queryValue, *user.Data.Age)
		argsPos++
	}

	if user.Data.Surname != nil {
		queryArgs = append(queryArgs, fmt.Sprintf("surname = &%d", argsPos))
		queryValue = append(queryValue, *user.Data.Surname)
		argsPos++
	}

	if user.Data.Email != "" {
		queryArgs = append(queryArgs, fmt.Sprintf("email = &%d", argsPos))
		queryValue = append(queryValue, user.Data.Email)
		argsPos++
	}

	if len(queryArgs) == 0 {
		return nil
	}

	queryArgs = append(queryArgs, fmt.Sprintf("updated_at = &%d", argsPos))
	queryValue = append(queryValue, time.Now())
	argsPos++

	queryText := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", tableName, strings.Join(queryArgs, ", "), argsPos)

	queryValue = append(queryValue, userId)

	bBuilder := builders.NewBatchBuilder()
	updateQ := bBuilder.AddExec()
	updateQ(queryText, queryValue...)

	return r.WithTxWithBatch(ctx, bBuilder)
}
func (r *repo) Create(ctx context.Context, user *models.User) (int64, error) {

	var newUserId int64
	password := user.Data.Password
	hexPassword, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if hashErr != nil {
		return 0, errors.New(fmt.Sprintf("error hashing password: %s", hashErr))
	}
	builder := builders.NewInsertBuilder(tableName)
	bBuilder := builders.NewBatchBuilder()

	builder.Set("name", user.Data.Name)
	builder.Set("email", user.Data.Email)
	builder.Set("password", string(hexPassword))
	builder.Set("updated_at", time.Now())

	if user.Data.Surname != nil {
		builder.Set("surname", *user.Data.Surname)
	}
	if user.Data.Age != nil {
		builder.Set("age", *user.Data.Age)
	}

	queryText, args := builder.Build()

	insertQ := bBuilder.AddQuery(&newUserId)
	insertQ(queryText, args)

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
