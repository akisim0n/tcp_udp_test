package user

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"tcp_upd_test/tcp/http/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeDB struct {
	beginFn func(context.Context) (pgx.Tx, error)
	queryFn func(context.Context, string, ...any) (pgx.Rows, error)
}

func (f *fakeDB) Begin(ctx context.Context) (pgx.Tx, error) {
	return f.beginFn(ctx)
}

func (f *fakeDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return f.queryFn(ctx, sql, args...)
}

type fakeTx struct {
	sendBatchFn func(context.Context, *pgx.Batch) pgx.BatchResults
	execFn      func(context.Context, string, ...any) (pgconn.CommandTag, error)
	committed   bool
	rolledBack  bool
}

func (f *fakeTx) Begin(context.Context) (pgx.Tx, error) { return nil, errors.New("not implemented") }
func (f *fakeTx) Commit(context.Context) error {
	f.committed = true
	return nil
}
func (f *fakeTx) Rollback(context.Context) error {
	f.rolledBack = true
	return nil
}
func (f *fakeTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, errors.New("not implemented")
}
func (f *fakeTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return f.sendBatchFn(ctx, b)
}
func (f *fakeTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }
func (f *fakeTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return f.execFn(ctx, sql, arguments...)
}
func (f *fakeTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeTx) QueryRow(context.Context, string, ...any) pgx.Row { return fakeRow{} }
func (f *fakeTx) Conn() *pgx.Conn                                  { return nil }

type fakeBatchResults struct {
	queryRowFn func() pgx.Row
	execFn     func() (pgconn.CommandTag, error)
}

func (f fakeBatchResults) Exec() (pgconn.CommandTag, error) {
	if f.execFn != nil {
		return f.execFn()
	}
	return pgconn.CommandTag{}, nil
}

func (f fakeBatchResults) Query() (pgx.Rows, error) { return nil, errors.New("not implemented") }

func (f fakeBatchResults) QueryRow() pgx.Row {
	if f.queryRowFn != nil {
		return f.queryRowFn()
	}
	return fakeRow{}
}

func (f fakeBatchResults) Close() error { return nil }

type fakeRow struct {
	values []any
	err    error
}

func (f fakeRow) Scan(dest ...any) error {
	if f.err != nil {
		return f.err
	}

	for i := range dest {
		if err := assignScanDest(dest[i], f.values[i]); err != nil {
			return err
		}
	}
	return nil
}

type fakeRows struct {
	rows    [][]any
	idx     int
	closed  bool
	scanErr error
	err     error
}

func (f *fakeRows) Close() { f.closed = true }
func (f *fakeRows) Err() error {
	return f.err
}
func (f *fakeRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}
func (f *fakeRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (f *fakeRows) Next() bool {
	if f.idx >= len(f.rows) {
		f.closed = true
		return false
	}
	f.idx++
	return true
}
func (f *fakeRows) Scan(dest ...any) error {
	if f.scanErr != nil {
		return f.scanErr
	}
	row := f.rows[f.idx-1]
	for i := range dest {
		if err := assignScanDest(dest[i], row[i]); err != nil {
			return err
		}
	}
	return nil
}
func (f *fakeRows) Values() ([]any, error) { return nil, errors.New("not implemented") }
func (f *fakeRows) RawValues() [][]byte    { return nil }
func (f *fakeRows) Conn() *pgx.Conn        { return nil }

func assignScanDest(dest any, value any) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("destination must be a non-nil pointer")
	}

	val := reflect.ValueOf(value)
	target := rv.Elem()

	if !val.IsValid() {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}

	if val.Type().AssignableTo(target.Type()) {
		target.Set(val)
		return nil
	}

	if val.Type().ConvertibleTo(target.Type()) {
		target.Set(val.Convert(target.Type()))
		return nil
	}

	return errors.New("cannot assign scanned value")
}

func TestCreate(t *testing.T) {
	testCases := []struct {
		name        string
		user        *models.User
		wantID      int64
		wantArgsLen int
	}{
		{
			name: "minimal fields",
			user: &models.User{
				Name:      "Alice",
				Email:     "alice@example.com",
				Password:  "hashed",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantID:      11,
			wantArgsLen: 5,
		},
		{
			name: "optional fields included",
			user: func() *models.User {
				surname := "Doe"
				age := int64(20)
				return &models.User{
					Name:      "Alice",
					Email:     "alice@example.com",
					Password:  "hashed",
					Surname:   &surname,
					Age:       &age,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
			}(),
			wantID:      11,
			wantArgsLen: 7,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tx := &fakeTx{
				sendBatchFn: func(_ context.Context, batch *pgx.Batch) pgx.BatchResults {
					if len(batch.QueuedQueries) != 1 {
						t.Fatalf("queued queries = %d, want 1", len(batch.QueuedQueries))
					}

					query := batch.QueuedQueries[0]
					if !strings.Contains(query.SQL, "INSERT INTO users") {
						t.Fatalf("unexpected query: %s", query.SQL)
					}

					if !strings.Contains(query.SQL, "RETURNING id") {
						t.Fatalf("query has no RETURNING id: %s", query.SQL)
					}

					if len(query.Arguments) != tc.wantArgsLen {
						t.Fatalf("args len = %d, want %d", len(query.Arguments), tc.wantArgsLen)
					}

					return fakeBatchResults{
						queryRowFn: func() pgx.Row {
							return fakeRow{values: []any{tc.wantID}}
						},
					}
				},
				execFn: func(context.Context, string, ...any) (pgconn.CommandTag, error) {
					return pgconn.CommandTag{}, nil
				},
			}

			repository := &repo{
				DB: &fakeDB{
					beginFn: func(context.Context) (pgx.Tx, error) { return tx, nil },
					queryFn: func(context.Context, string, ...any) (pgx.Rows, error) {
						return nil, errors.New("unexpected query")
					},
				},
			}

			id, err := repository.Create(context.Background(), tc.user)
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}

			if id != tc.wantID {
				t.Fatalf("Create() id = %d, want %d", id, tc.wantID)
			}

			if !tx.committed {
				t.Fatal("transaction was not committed")
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	updatedAt := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	testCases := []struct {
		name     string
		user     *models.User
		wantSQL  string
		wantArgs []any
	}{
		{
			name: "changed name and email",
			user: &models.User{
				ID:        5,
				Name:      "Alice",
				Email:     "alice@example.com",
				UpdatedAt: updatedAt,
			},
			wantSQL:  "UPDATE users SET name = $1, email = $2, updated_at = $3 WHERE id = $4",
			wantArgs: []any{"Alice", "alice@example.com", updatedAt, int64(5)},
		},
		{
			name: "only updated at",
			user: &models.User{
				ID:        5,
				UpdatedAt: updatedAt,
			},
			wantSQL:  "UPDATE users SET updated_at = $1 WHERE id = $2",
			wantArgs: []any{updatedAt, int64(5)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tx := &fakeTx{
				sendBatchFn: func(_ context.Context, batch *pgx.Batch) pgx.BatchResults {
					if len(batch.QueuedQueries) != 1 {
						t.Fatalf("queued queries = %d, want 1", len(batch.QueuedQueries))
					}

					query := batch.QueuedQueries[0]
					if query.SQL != tc.wantSQL {
						t.Fatalf("sql = %q, want %q", query.SQL, tc.wantSQL)
					}

					if !reflect.DeepEqual(query.Arguments, tc.wantArgs) {
						t.Fatalf("args = %#v, want %#v", query.Arguments, tc.wantArgs)
					}

					return fakeBatchResults{
						execFn: func() (pgconn.CommandTag, error) { return pgconn.CommandTag{}, nil },
					}
				},
				execFn: func(context.Context, string, ...any) (pgconn.CommandTag, error) {
					return pgconn.CommandTag{}, nil
				},
			}

			repository := &repo{
				DB: &fakeDB{
					beginFn: func(context.Context) (pgx.Tx, error) { return tx, nil },
					queryFn: func(context.Context, string, ...any) (pgx.Rows, error) {
						return nil, errors.New("unexpected query")
					},
				},
			}

			err := repository.Update(context.Background(), tc.user)
			if err != nil {
				t.Fatalf("Update() error = %v", err)
			}
		})
	}
}

func TestGetAll(t *testing.T) {
	testCases := []struct {
		name      string
		rows      [][]any
		wantCount int
		checkFn   func(*testing.T, []*models.User)
	}{
		{
			name: "scans multiple users",
			rows: func() [][]any {
				surname := "Doe"
				age1 := int64(20)
				age2 := int64(30)
				return [][]any{
					{int64(1), "Alice", &surname, &age1, time.Time{}, time.Time{}, "alice@example.com"},
					{int64(2), "Bob", (*string)(nil), &age2, time.Time{}, time.Time{}, "bob@example.com"},
				}
			}(),
			wantCount: 2,
			checkFn: func(t *testing.T, users []*models.User) {
				if users[0].Email != "alice@example.com" || users[1].Name != "Bob" {
					t.Fatalf("unexpected users: %+v", users)
				}
			},
		},
		{
			name:      "empty result",
			rows:      [][]any{},
			wantCount: 0,
			checkFn:   func(*testing.T, []*models.User) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := &fakeRows{rows: tc.rows}
			repository := &repo{
				DB: &fakeDB{
					beginFn: func(context.Context) (pgx.Tx, error) { return nil, errors.New("unexpected begin") },
					queryFn: func(_ context.Context, sql string, _ ...any) (pgx.Rows, error) {
						if !strings.Contains(sql, "SELECT id, name, surname") {
							t.Fatalf("unexpected query: %s", sql)
						}
						return rows, nil
					},
				},
			}

			users, err := repository.GetAll(context.Background())
			if err != nil {
				t.Fatalf("GetAll() error = %v", err)
			}

			if len(users) != tc.wantCount {
				t.Fatalf("len(users) = %d, want %d", len(users), tc.wantCount)
			}

			tc.checkFn(t, users)
		})
	}
}

func TestDelete(t *testing.T) {
	testCases := []struct {
		name string
		id   int64
	}{
		{name: "positive id", id: 9},
		{name: "different id", id: 42},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tx := &fakeTx{
				sendBatchFn: func(context.Context, *pgx.Batch) pgx.BatchResults {
					return fakeBatchResults{}
				},
				execFn: func(_ context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
					if sql != "DELETE FROM users WHERE id = $1" {
						t.Fatalf("sql = %q", sql)
					}

					if !reflect.DeepEqual(arguments, []any{tc.id}) {
						t.Fatalf("args = %#v, want %#v", arguments, []any{tc.id})
					}

					return pgconn.CommandTag{}, nil
				},
			}

			repository := &repo{
				DB: &fakeDB{
					beginFn: func(context.Context) (pgx.Tx, error) { return tx, nil },
					queryFn: func(context.Context, string, ...any) (pgx.Rows, error) {
						return nil, errors.New("unexpected query")
					},
				},
			}

			if err := repository.Delete(context.Background(), tc.id); err != nil {
				t.Fatalf("Delete() error = %v", err)
			}

			if !tx.committed {
				t.Fatal("transaction was not committed")
			}
		})
	}
}
