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

type queryData struct {
	Text string
	Args []interface{}
}

func NewRepository(db *pgxpool.Pool) repository.UserRepository {
	return &repo{DB: db}
}

func (r *repo) Get(ctx context.Context, id int64) (*models.User, error) {

	var user models.User

	execErr := r.DB.QueryRow(ctx, "SELECT * FROM $1 WHERE id = $2", tableName, id).Scan(&user)
	if execErr != nil {
		return nil, execErr
	}

	return &user, nil
}
func (r *repo) GetAll(ctx context.Context) ([]*models.User, error) {

	rowData, execErr := r.DB.Query(ctx, "SELECT * FROM users")
	if execErr != nil {
		return nil, execErr
	}
	defer rowData.Close()

	var users []*models.User

	for rowData.Next() {
		var user models.User
		if scanErr := rowData.Scan(&user); scanErr != nil {
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

	updateQuery := queryData{
		Text: queryText,
		Args: queryValue,
	}

	return r.ExecUnderTransaction(ctx, updateQuery)
}
func (r *repo) Create(ctx context.Context, user *models.User) (int64, error) {

	password := user.Data.Password
	hexPassword, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if hashErr != nil {
		return 0, errors.New(fmt.Sprintf("error hashing password: %s", hashErr))
	}
	builder := builders.NewInsertBuilder(tableName)

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

	execErr := r.ExecUnderTransaction(ctx, queryData{Text: queryText, Args: args})
	if execErr != nil {
		return 0, execErr
	}

	return 0, nil
}
func (r *repo) Delete(ctx context.Context, id int64) error {

	queryText := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName)

	return r.ExecUnderTransaction(ctx, queryData{Text: queryText, Args: []interface{}{id}})
}

func (r *repo) ExecUnderTransaction(ctx context.Context, queryArgs ...queryData) error {
	trans, err := r.DB.Begin(ctx)
	if err != nil {
		return errors.New(fmt.Sprintf("error starting transaction: %s", err))
	}

	batch := &pgx.Batch{}

	for _, queryArg := range queryArgs {
		batch.Queue(queryArg.Text, queryArg.Args...)
	}

	res := trans.SendBatch(ctx, batch)

	for i := 0; i < batch.Len(); i++ {
		_, execErr := res.Exec()
		if execErr != nil {
			for j := i + 1; j < batch.Len(); j++ {
				_, _ = res.Exec()
			}
			res.Close()
			trans.Rollback(ctx)
			return errors.New(fmt.Sprintf("error executing batch result: %s", execErr))
		}
	}

	resCloseErr := res.Close()
	if resCloseErr != nil {
		trans.Rollback(ctx)
		return errors.New(fmt.Sprintf("error closing batch result: %s", resCloseErr))
	}
	commitErr := trans.Commit(ctx)
	if commitErr != nil {
		return errors.New(fmt.Sprintf("error commiting transaction: %s", commitErr))
	}
	return nil
}
