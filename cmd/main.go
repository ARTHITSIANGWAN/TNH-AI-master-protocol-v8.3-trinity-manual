package main

import (
  "fmt"
  "os"
  "net/http"
)

/**
 * 🛡️ TNH AI V83 TRINITY EMPIRE
 * Zero-Garbage Logic: 100% Pure Go
 * จัดย่อหน้าแบบ 2 Spaces ตามมาตรฐานที่บอสเลือก
 */

func main() {
  // ดึงค่าจาก Environment เพื่อความปลอดภัย (Security Check)
  port := os.Getenv("PORT")
  if port == "" {
    port = "2026" 
  }

  // ฟังก์ชันตรวจสอบสถานะขุนพล
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "🐯 V83 TRINITY: SYSTEM ACTIVE (0.xxms)")
  })

  // Ignite Engine
  fmt.Printf("🚀 V83 Engine เริ่มทำงานที่พอร์ต %s\n", port)
  if err := http.ListenAndServe(":"+port, nil); err != nil {
    fmt.Printf("❌ Error: %v\n", err)
  }
}
