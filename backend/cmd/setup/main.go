package main

import (
	"bufio"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "cannot initialize .env:", err)
		os.Exit(1)
	}
}

func run() error {
	path := "../.env"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	values := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))

	for scanner.Scan() {
		k, v, ok := strings.Cut(scanner.Text(), "=")
		if ok {
			values[k] = v
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	keys, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	if (values["VAPID_PRIVATE_KEY"] == "") != (values["VAPID_PUBLIC_KEY"] == "") {
		return fmt.Errorf("both VAPID keys must be present or empty")
	}

	defaults := map[string]string{"FRONTEND_PORT": "3000", "POSTGRES_PASSWORD": rand.Text(), "API_DB_PASSWORD": rand.Text(), "WORKER_DB_PASSWORD": rand.Text(), "NOTIFICATIONS_DB_PASSWORD": rand.Text(), "NOTIFICATIONS_TOKEN": rand.Text(), "JWT_SECRET": rand.Text() + rand.Text(), "VAPID_PRIVATE_KEY": base64.RawURLEncoding.EncodeToString(keys.Bytes()), "VAPID_PUBLIC_KEY": base64.RawURLEncoding.EncodeToString(keys.PublicKey().Bytes()), "VAPID_SUBJECT": "mailto:admin@example.com", "APP_ORIGINS": "http://localhost:3000,http://localhost:5173"}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")

	for key, value := range defaults {
		if values[key] != "" {
			continue
		}

		found := false

		for i, line := range lines {
			if strings.HasPrefix(line, key+"=") {
				lines[i] = key + "=" + value
				found = true
				break
			}
		}

		if !found {
			lines = append(lines, key+"="+value)
		}
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600)
}
