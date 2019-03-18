package dotenv_test

import (
	"os"
	"testing"

	"github.com/silvernemesis/jwtproject/internal/dotenv"
)

var results = map[string]string{
	"ADMIN_USERNAME": "admin",
	"ADMIN_PASSWORD": "password",
	"SECRET_KEY":     "secret",
	"CERT_FILE":      "server.crt",
	"KEY_FILE":       "server.key",
}

var input = []string{
	"ADMIN_USERNAME=admin",
	"ADMIN_PASSWORD=password",
	"SECRET_KEY=secret",
	"CERT_FILE=server.crt",
	"KEY_FILE=server.key",
}

func TestDotEnv(t *testing.T) {
	err := dotenv.ProcessFile(".env")
	if err != nil {
		t.Error("error loading .env file")
	}
	for key := range results {
		value := os.Getenv(key)

		if value != results[key] {
			t.Error("expected", results[key], "got", value)
		}
	}
}

func BenchmarkDotEnv(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dotenv.ParseLines(input)
	}
}
