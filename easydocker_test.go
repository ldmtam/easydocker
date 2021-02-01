package easydocker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

var pool *Pool

func TestMain(m *testing.M) {
	var err error
	pool, err = NewPool("")
	if err != nil {
		panic(err)
	}

	exitCode := m.Run()

	if err := pool.Close(); err != nil {
		panic(err)
	}

	os.Exit(exitCode)
}

func TestMySQL(t *testing.T) {
	containerID, err := pool.CreateContainer(
		"mysql",
		WithEnvironment(
			"MYSQL_ROOT_PASSWORD=example",
			"MYSQL_DATABASE=test",
		),
		WithExposedPorts("3306"),
	)
	assert.Nil(t, err)
	if len(containerID) == 0 {
		t.Fatal("container id can't be empty")
	}

	printContainerInfo(containerID)

	hostport, err := pool.GetHostPort(containerID, "3306")
	assert.Nil(t, err)
	t.Logf("Docker Port: %v => Hostport: %v\n", "3306", hostport)

	err = pool.Retry(context.Background(), func() error {
		db, err := sql.Open("mysql",
			fmt.Sprintf("root:example@tcp(%v)/test?charset=utf8mb4&parseTime=True&loc=Local", hostport),
		)
		if err != nil {
			return err
		}
		defer db.Close()

		return db.Ping()
	})
	assert.Nil(t, err)
}

func TestMySQLWithMount(t *testing.T) {
	containerID, err := pool.CreateContainer(
		"mysql",
		WithEnvironment("MYSQL_ROOT_PASSWORD=example"),
		WithMounts("/var/lib/mysql"),
	)
	assert.Nil(t, err)
	if len(containerID) == 0 {
		t.Fatal("container id can't be empty")
	}

	printContainerInfo(containerID)
}

func TestMySQLWithCmd(t *testing.T) {
	containerID, err := pool.CreateContainer(
		"mysql",
		WithCmd(
			"--default-authentication-plugin=mysql_native_password",
			"--character-set-server=utf8mb4",
			"--collation-server=utf8mb4_unicode_ci",
		),
		WithEnvironment("MYSQL_ROOT_PASSWORD=example"),
	)
	assert.Nil(t, err)
	if len(containerID) == 0 {
		t.Fatal("container id can't be empty")
	}

	printContainerInfo(containerID)
}

func TestMySQLWithNetwork(t *testing.T) {
	networkID, err := pool.CreateNetwork("my-custom-network")
	assert.Nil(t, err)

	containerID, err := pool.CreateContainer(
		"mysql",
		WithEnvironment("MYSQL_ROOT_PASSWORD=example"),
		WithNetwork(networkID),
	)
	assert.Nil(t, err)

	printContainerInfo(containerID)
}

func printContainerInfo(containerID string) {
	containerInfo, err := pool.GetContainerInfo(containerID)
	if err != nil {
		panic(err)
	}
	out := map[string]interface{}{
		"id":            containerInfo.ID,
		"created_at":    containerInfo.Created,
		"status":        containerInfo.State.Status,
		"image":         containerInfo.Config.Image,
		"exposed_ports": containerInfo.Config.ExposedPorts,
		"mount": map[string]string{
			"source":      containerInfo.Mounts[0].Source,
			"destination": containerInfo.Mounts[0].Destination,
		},
		"cmds": containerInfo.Config.Cmd,
	}
	log.Println("Container Info:", spew.Sdump(out))
}
