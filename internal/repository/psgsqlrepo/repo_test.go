package psgsqlrepo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresTestSuite struct {
	suite.Suite
	psgsqlRepo

	tc *tcpostgres.PostgresContainer
}

func (r *psgsqlRepo) cleanTables(ctx context.Context) error {
	query := `DELETE FROM users_shorten_url`
	_, err := r.conn.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	query = `DELETE FROM users`
	_, err = r.conn.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	query = `DELETE FROM shorten_url`
	_, err = r.conn.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return err
}

// SetupSuite инициализирует PostgresTestSuite
func (ts *PostgresTestSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pgc, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:16.2"),
		tcpostgres.WithDatabase("shortenerdb"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("123"),
		tcpostgres.WithInitScripts(),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(ts.T(), err)
	ts.tc = pgc

	dbConnectionString, err := getDBConnectionString(ctx, pgc, "shortenerdb", "123")
	require.NoError(ts.T(), err)
	db, err := NewSQLxConnection(ctx, dbConnectionString)
	require.NoError(ts.T(), err)
	repository, err := NewPsgsqlRepo(ctx, db)
	require.NoError(ts.T(), err)
	ts.psgsqlRepo = *repository

	ts.T().Logf("stared postgres conteiner")
}

// TearDownSuite останавливает контейнеры
func (ts *PostgresTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(ts.T(), ts.tc.Terminate(ctx))
}

// SetupTest очищает БД
func (ts *PostgresTestSuite) SetupTest() {
	ts.Require().NoError(ts.cleanTables(context.Background()))
}

// TearDownTest очищает БД
func (ts *PostgresTestSuite) TearDownTest() {
	ts.Require().NoError(ts.cleanTables(context.Background()))
}

// TestPostgresqlRepository входная точка для тестирования
func TestPostgresqlRepository(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}

func getDBConnectionString(
	ctx context.Context,
	pgc *tcpostgres.PostgresContainer,
	dbname string,
	password string) (string, error) {
	host, err := pgc.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := pgc.MappedPort(ctx, "5432")
	if err != nil {
		return "", err
	}
	portDigit := uint16(port.Int())

	return fmt.Sprintf(
		"host=%s user=postgres password=%s dbname=%s port=%d sslmode=disable",
		host, password, dbname, portDigit), nil
}
