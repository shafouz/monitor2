package database

import (
	"context"
	"log"
	"monitor2/src/db/models"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB Database

type Database struct {
	Pool *pgxpool.Pool
}

func Init() error {
	url := os.Getenv("DATABASE_URL")
	if len(url) == 0 {
		log.Fatal("DATABASE_URL is empty")
	}

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		log.Fatal(err)
	}

	DB.Pool = pool
	err = DB.Pool.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.New("file://db/migrations/", url)
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	return nil
}

func (db Database) GetManyEndpointsBySchedule(schedule int) ([]models.Endpoint, error) {
	rows, err := db.Pool.Query(context.Background(), "SELECT * FROM Endpoint WHERE schedule_hours = $1", schedule)
	if err != nil {
		return []models.Endpoint{}, err
	}
	r, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Endpoint])
	if err != nil {
		return []models.Endpoint{}, err
	}
	return r, nil
}

func (db Database) GetEndpointByUrl(url string) (models.Endpoint, error) {
	rows, err := db.Pool.Query(context.Background(), "SELECT * FROM Endpoint WHERE url = $1", url)
	if err != nil {
		return models.Endpoint{}, err
	}
	r, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Endpoint])
	if err != nil {
		return models.Endpoint{}, err
	}
	return r, nil
}

func (db Database) CreateEndpoint(endpoint models.Endpoint) error {
	_, err := db.Pool.Exec(context.Background(),
		`INSERT INTO Endpoint ( url, status_code, response_body, previous_response_body, selector, profile )
    VALUES ( $1, $2, $3, $4, $5, $6 )`,
		endpoint.Url,
		endpoint.StatusCode,
		endpoint.ResponseBody,
		endpoint.PreviousResponseBody,
		endpoint.Selector,
		endpoint.Profile,
	)
	if err != nil {
		return err
	}
	return nil
}

func (db Database) UpdateEndpointByUrl(endpoint models.Endpoint, from_crawler bool) error {
	if from_crawler {
		_, err := db.Pool.Exec(context.Background(),
			`UPDATE Endpoint
      SET url = $1,
      status_code = $2,
      response_body = $3,
      previous_response_body = $4,
      selector = $5,
      profile = $6
      WHERE url = $1`,
			endpoint.Url,
			endpoint.StatusCode,
			endpoint.ResponseBody,
			endpoint.PreviousResponseBody,
			endpoint.Selector,
			endpoint.Profile,
		)
		if err != nil {
			return err
		}
		return nil
	}

	_, err := db.Pool.Exec(context.Background(),
		`UPDATE Endpoint
      SET url = $1,
      selector = $2,
      profile = $3,
      deleted = $4
      WHERE url = $1`,
		endpoint.Url,
		endpoint.Selector,
		endpoint.Profile,
    endpoint.Deleted,
	)
	if err != nil {
		return err
	}
	return nil
}

func (db Database) DeleteEndpoint(endpoint models.Endpoint) (int, error) {
	t, err := db.Pool.Exec(context.Background(),
		"DELETE FROM Endpoint WHERE url = $1",
		endpoint.Url,
	)
	if err != nil {
		return 0, err
	}
	return int(t.RowsAffected()), nil
}

func (db Database) GetAllEndpoints() ([]models.Endpoint, error) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT * FROM Endpoint`,
	)
	if err != nil {
		return nil, err
	}

	diff, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Endpoint])
	if err != nil {
		return nil, err
	}
	return diff, nil
}

func (db Database) GetAllUrls() ([]string, error) {
	rows, err := db.Pool.Query(context.Background(), "SELECT Url FROM Endpoint")
	if err != nil {
		return []string{}, err
	}

	r, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return []string{}, err
	}
	return r, nil
}

func (db Database) GetAllRepos() ([]models.Repository, error) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT * FROM Repository`,
	)
	if err != nil {
		return nil, err
	}

	diff, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Repository])
	if err != nil {
		return nil, err
	}
	return diff, nil
}

func (db Database) GetManyRepositoriesBySchedule(schedule int) ([]models.Repository, error) {
	rows, err := db.Pool.Query(
		context.Background(),
		"SELECT * FROM Repository WHERE schedule_hours = $1",
		schedule,
	)

	if err != nil {
		return []models.Repository{}, err
	}

	r, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Repository])
	if err != nil {
		return []models.Repository{}, err
	}

	return r, nil
}

func (db Database) UpdateRepository(id int, repository models.Repository) error {
	_, err := db.Pool.Exec(context.Background(),
		`UPDATE Repository 
    SET url = $2,
    directory = $3,
    watched_files = $4,
    remote = $5,
    schedule_hours = $6,
    deleted = $7
    WHERE id = $1`,
		id,
		repository.Url,
		repository.Directory,
		repository.WatchedFiles,
		repository.Remote,
		repository.ScheduleHours,
		repository.Deleted,
	)
	if err != nil {
		return err
	}
	return nil
}

func (db Database) CreateRepository(repository models.Repository) error {
	_, err := db.Pool.Exec(context.Background(),
		`INSERT INTO Repository ( url, directory, watched_files, remote )
    VALUES ( $1, $2, $3, $4 )`,
		repository.Url,
		repository.Directory,
		repository.WatchedFiles,
		repository.Remote,
	)
	if err != nil {
		return err
	}
	return nil
}

func (db Database) GetDiff(id string) (models.Diff, error) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT id, body, url, commit, created_at FROM Diff WHERE id = $1`,
		id,
	)
	if err != nil {
		return models.Diff{}, err
	}

	diff, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Diff])
	if err != nil {
		return models.Diff{}, err
	}
	return diff, nil
}

func (db Database) GetAllDiffs() ([]models.Diff, error) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT id, url, '' as body, commit, created_at FROM Diff ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}

	diff, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Diff])
	if err != nil {
		return nil, err
	}
	return diff, nil
}

func (db Database) CreateDiff(diff models.Diff) error {
	_, err := db.Pool.Exec(context.Background(),
		`INSERT INTO Diff ( id, body, url, commit )
    VALUES ( $1, $2, $3, $4 )`,
		diff.Id,
		diff.Body,
		diff.Url,
    diff.Commit,
	)
	if err != nil {
		return err
	}
	return nil
}
