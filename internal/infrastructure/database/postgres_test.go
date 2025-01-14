package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDatabase(t *testing.T) (string, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:14",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatal(err)
	}

	dbURL := fmt.Sprintf("postgres://test:test@localhost:%s/testdb?sslmode=disable", port.Port())

	// Função de cleanup
	cleanup := func() {
		container.Terminate(ctx)
	}

	return dbURL, cleanup
}

func TestUserRepository(t *testing.T) {
	_, cleanup := setupTestDatabase(t)
	defer cleanup()

	// Execute as migrações no banco de teste
	// Implemente seus testes aqui
}
