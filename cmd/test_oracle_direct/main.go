package main

import (
	"context"
	"fmt"
	"time"

	_ "github.com/sijms/go-ora/v2"
	"database/sql"
)

func main() {
	// Test different DSN formats
	testDSNs := []string{
		"system/Qwer1234@//192.168.170.137:1521/orcl",
		"system/Qwer1234@192.168.170.137:1521/orcl",
		"oracle://system:Qwer1234@192.168.170.137:1521/orcl",
	}

	for i, dsn := range testDSNs {
		fmt.Printf("\n[Test %d] DSN: %s\n", i+1, dsn)
		
		db, err := sql.Open("oracle", dsn)
		if err != nil {
			fmt.Printf("  ❌ Open failed: %v\n", err)
			continue
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err = db.PingContext(ctx)
		cancel()
		
		if err != nil {
			fmt.Printf("  ❌ Ping failed: %v\n", err)
		} else {
			fmt.Printf("  ✅ Success!\n")
			
			// Get version
			var version string
			err = db.QueryRowContext(context.Background(), "SELECT * FROM v$version WHERE rownum = 1").Scan(&version)
			if err == nil {
				fmt.Printf("  Version: %s\n", version)
			}
		}
		db.Close()
	}
}
