package redisstore

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"testing"
	"time"

	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/redis/go-redis/v9"
)

var _ port.ActiveSessionReader = NewActiveSessionReader(nil)

func TestActiveSessionReaderWithoutClientFailsClosed(t *testing.T) {
	reader := NewActiveSessionReader(nil)
	if _, err := reader.IsRevoked(context.Background(), "s"); err == nil {
		t.Fatal("sem cliente precisa dar erro")
	}
	if _, err := reader.Entries(context.Background(), "s"); err == nil {
		t.Fatal("sem cliente precisa dar erro")
	}
}

func TestActiveSessionReaderRedisDownIsError(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", DialTimeout: 100 * time.Millisecond, MaxRetries: -1})
	defer client.Close()
	reader := NewActiveSessionReader(client)
	if _, err := reader.IsRevoked(context.Background(), "s"); err == nil {
		t.Fatal("Redis fora precisa dar erro")
	}
	if _, err := reader.Entries(context.Background(), "s"); err == nil {
		t.Fatal("Redis fora precisa dar erro")
	}
}

// TestActiveSessionReaderAgainstRedis roda contra redis-server temporário, com as chaves
// e o JSON como o ms-auth grava. Sem redis-server na máquina, não roda.
func TestActiveSessionReaderAgainstRedis(t *testing.T) {
	path, err := exec.LookPath("redis-server")
	if err != nil {
		t.Skip("redis-server não instalado")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	server := exec.Command(path, "--port", fmt.Sprint(port), "--save", "", "--appendonly", "no", "--dir", t.TempDir())
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Process.Kill(); _ = server.Wait() })

	client := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("127.0.0.1:%d", port)})
	defer client.Close()
	ctx := context.Background()
	for i := 0; ; i++ {
		if client.Ping(ctx).Err() == nil {
			break
		}
		if i > 100 {
			t.Fatal("redis-server não subiu")
		}
		time.Sleep(20 * time.Millisecond)
	}

	reader := NewActiveSessionReader(client)
	entries, err := reader.Entries(ctx, "inventado")
	if err != nil || len(entries) != 0 {
		t.Fatalf("sid inventado: %v %v", entries, err)
	}
	client.SAdd(ctx, "refreshtoken:sid:sid-1", "h1", "expirado", "ilegivel")
	client.Set(ctx, "refreshtoken:h1", `{"tokenHash":"h1","sid":"sid-1","codeUser":"u1","companyId":"c1","clientId":"web","deviceId":null,"familyId":"f","createdAt":"x","absoluteExpiresAt":"y","replacedByHash":null,"revoked":false}`, time.Minute)
	client.Set(ctx, "refreshtoken:ilegivel", `{nao json`, time.Minute)
	entries, err = reader.Entries(ctx, "sid-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].CodeUser != "u1" || entries[0].CompanyID != "c1" || entries[0].SID != "sid-1" || entries[0].Revoked {
		t.Fatalf("entrada: %+v", entries)
	}
	revoked, err := reader.IsRevoked(ctx, "sid-1")
	if err != nil || revoked {
		t.Fatalf("revogada: %v %v", revoked, err)
	}
	client.Set(ctx, "session:revoked:sid-1", "1", time.Minute)
	revoked, err = reader.IsRevoked(ctx, "sid-1")
	if err != nil || !revoked {
		t.Fatalf("revogada: %v %v", revoked, err)
	}
}
