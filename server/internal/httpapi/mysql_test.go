package httpapi

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net"
	"os"
	"strconv"
	"testing"

	"myself/server/internal/store"

	"github.com/go-sql-driver/mysql"
)

// Run the existing HTTP regression suite against isolated MySQL databases when
// MYSELF_TEST_MYSQL_ADDR is set. Never reuse or clear an operator's database.
func openTestDB(t *testing.T, dataDir string) (*store.DB, error) {
	t.Helper()
	addr := os.Getenv("MYSELF_TEST_MYSQL_ADDR")
	if addr == "" {
		return store.Open(dataDir)
	}
	host, portText, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatal(err)
	}
	port, _ := strconv.Atoi(portText)
	user := os.Getenv("MYSELF_TEST_MYSQL_USER")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("MYSELF_TEST_MYSQL_PASSWORD")
	config := mysql.NewConfig()
	config.Net, config.Addr, config.User, config.Passwd = "tcp", addr, user, password
	admin, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	name := "myself_test_" + hex.EncodeToString(random)
	if _, err := admin.Exec("CREATE DATABASE `" + name + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin"); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP DATABASE `" + name + "`"); err != nil {
			t.Error(err)
		}
		admin.Close()
	})
	return store.OpenConfigured(dataDir, store.DatabaseConfig{Driver: "mysql", MySQL: store.MySQLConfig{Host: host, Port: port, Name: name, User: user, Password: password}})
}
