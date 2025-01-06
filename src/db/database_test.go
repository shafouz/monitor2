package database_test

import (
	"context"
	"log"
	database "monitor2/src/db"
	"monitor2/src/db/models"
	"time"

	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var db_on bool = false
var DB database.Database
var ctx context.Context = context.Background()

func start_test_db() {
  if db_on {
    return
  }

	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal("Error creating Docker client:", err)
	}

	out, err := cli.ImagePull(ctx, "postgres", types.ImagePullOptions{})
	if err != nil {
		log.Fatal("Error pulling image:", err)
	}
	defer out.Close()

	cli.ContainerStop(context.Background(), "pg-test", container.StopOptions{})
	cli.ContainerRemove(context.Background(), "pg-test", types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})

	container_cfg := &container.Config{
		Healthcheck: &container.HealthConfig{
			Test: []string{
				"CMD",
				"pg_isready",
				"-U",
				"pg",
				"-d",
				"test",
			},
			Timeout:     time.Duration(10) * time.Second,
			Retries:     5,
			Interval:    time.Duration(2) * time.Second,
			StartPeriod: time.Duration(20) * time.Second,
		},
		Image: "postgres",
		ExposedPorts: nat.PortSet{
			"5432/tcp": struct{}{},
		},
		Env: []string{
			"POSTGRES_USER=pg",
			"POSTGRES_PASSWORD=test",
			"POSTGRES_HOST=0.0.0.0",
			"POSTGRES_DB=test",
		},
	}

	url := "postgres://pg:test@0.0.0.0:5433/test?sslmode=disable"

	host_cfg := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port("5432/tcp"): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "5433",
				},
			},
		},
	}

	resp, err := cli.ContainerCreate(ctx, container_cfg, host_cfg, nil, nil, "pg-test")
	if err != nil {
		log.Fatal("Error creating container:", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Fatal("Error starting container:", err)
	}

	for {
		container, err := cli.ContainerInspect(ctx, "pg-test")
		if err != nil {
			log.Fatal(err)
		}

		health := container.State.Health
		if health != nil && health.Status == "healthy" {
			log.Println("Container is healthy, you can now interact with it.")
			break
		} else {
			log.Println("Waiting for the container to be healthy...")
		}
		time.Sleep(2 * time.Second)
	}

	pool, err := pgxpool.New(ctx, url)
	DB.Pool = pool
	err = DB.Pool.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.New("file://../../db/migrations/", url)
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

  db_on = true
}

func clean_db() {
  DB.Pool.Exec(ctx, "SELECT 'TRUNCATE ' || table_name || ';' FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'")
}

func TestRepos(t *testing.T) {
  start_test_db()

	repo := models.Repository{
		Url:           "https://example.com/repo2",
		Directory:     "/path/to/repo2",
		WatchedFiles:  []byte(`["folder/file1.txt", "folder/file2.txt"]`),
		Remote:        "upstream",
		ScheduleHours: 48,
		Deleted:       true,
		UpdatedAt:     time.Now().Add(-24 * time.Hour),
	}

	repo2 := models.Repository{
		Url:           "https://alallala.com/repo2",
		Directory:     "/path/toa",
		WatchedFiles:  []byte(`["ile1.txt", "folder/file2.txt"]`),
		Remote:        "origin",
		ScheduleHours: 51,
		Deleted:       true,
		UpdatedAt:     time.Now().Add(-5 * time.Hour),
	}

	err := DB.CreateRepository(repo)
	if err != nil {
		t.Fatal(err)
	}

	err = DB.UpdateRepository(1, repo2)
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}

  clean_db()
}

func TestCreatesAndGetsDiff(t *testing.T) {
  start_test_db()

	id := uuid.New().String()
	diff := models.Diff{
		Id:     id,
		Body:   "lalalala",
		Url:    "asdsadsad",
		Commit: "asdsad",
	}
	err := DB.CreateDiff(diff)
	if err != nil {
		t.Fatal(err)
	}

	_, err = DB.GetDiff(id)
	if err != nil {
		t.Fatal(err)
	}

	_, err = DB.GetAllDiffs()
	if err != nil {
		t.Fatal(err)
	}

  clean_db()
}
