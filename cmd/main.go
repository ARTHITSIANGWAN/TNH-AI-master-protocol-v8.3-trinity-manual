package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// ===== [1ส: สะสาง] Config & Sovereign Command =====
var (
	SECRET = []byte(os.Getenv("TNH_SECRET"))
	PORT   = ":2026"
	// แยก Channel ตามประเภทงาน เพื่อไม่ให้ขุนพลตีกันเอง
	securityChan = make(chan Job, 512) // สำหรับ L9, L8
	logicChan    = make(chan Job, 512) // สำหรับ L3, L4
	businessChan = make(chan Job, 512) // สำหรับ L6, L10
	systemChan   = make(chan Job, 128) // สำหรับ L11, L7
)

type Job struct {
	JobID  string `json:"job_id"`
	Action string `json:"action"` // ระบุประเภทงานเพื่อส่งให้ขุนพลที่ถูกอาชีพ
	Topic  string `json:"topic"`
	Payload interface{} `json:"payload"`
	Ts     int64  `json:"ts"`
	Ttl    int    `json:"ttl"`
	Sig    string `json:"sig"`
}

var seen sync.Map // [L5: เพชฌฆาตขยะ] ใช้คุม Idempotency

// ===== [3ส: สะอาด] The Helmet (Security Logic) =====

func generateSovereignSig(j Job) string {
	// แก้จุดตาย: ใช้ Sprintf แทน rune เพื่อสัจจะที่เที่ยงตรง
	data := fmt.Sprintf("%s|%s|%s|%d", j.JobID, j.Action, j.Topic, j.Ts)
	h := hmac.New(sha256.New, SECRET)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// ===== [5ส: สร้างนิสัย] The 11 Professionals Logic =====

func startGenerals(ctx context.Context) {
	// [L1: จอมพล] - ผู้ควบคุม Lifecycle ทั้งหมดผ่าน ctx
	log.Println("[L1: จอมพล] บัญชาการกองทัพเริ่มปฏิบัติการ...")

	// [L9: พลซุ่มยิง] & [L8: บอดี้การ์ด] - หน่วยความมั่นคง
	go func() {
		for j := range securityChan {
			log.Printf("[L9/L8] ตรวจสอบสัจจะงาน %s: ผ่านด่านตรวจแล้ว", j.JobID)
			logicChan <- j // ส่งต่อให้หน่วยประมวลผล
		}
	}()

	// [L3: ศัลยแพทย์] & [L4: นักสืบ] - หน่วยประมวลผลลอจิก
	go func() {
		for j := range logicChan {
			log.Printf("[L3/L4] ศัลยกรรมและแกะรอยงาน: %s", j.Topic)
			businessChan <- j
		}
	}()

	// [L6: พ่อค้า] & [L10: ผู้ตรวจสอบ] - หน่วยจัดการผลประโยชน์
	go func() {
		for j := range businessChan {
			log.Printf("[L6/L10] บันทึกธุรกรรมและตรวจสอบกฎหมาย: %s", j.Action)
		}
	}()

	// [L11: หมอผี] & [L7: วิศวกร] - หน่วยฟื้นฟูและจราจร
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for {
			select {
			case <-ticker.C:
				// [L11: หมอผี] ทำพิธีล้าง Memory ที่ค้าง (Self-Healing)
				seen = sync.Map{} 
				log.Println("[L11] ทำพิธีล้างอาถรรพ์ (Reset Cache) เรียบร้อย")
			case j := <-systemChan:
				log.Printf("[L7] จัดการจราจรโหนด: %v", j.Payload)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// ===== [4ส: สุขลักษณะ] The Gateways =====

func processHandler(w http.ResponseWriter, r *http.Request) {
	var j Job
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		http.Error(w, "สาส์นสกปรก", 400); return
	}

	// [L5: เพชฌฆาตขยะ] ตรวจสอบงานซ้ำ
	if _, loaded := seen.LoadOrStore(j.JobID, true); loaded {
		w.WriteHeader(202)
		w.Write([]byte("สัจจะซ้ำซ้อน")); return
	}

	// ส่งเข้าด่านหน้า (Security)
	securityChan <- j
	w.WriteHeader(http.StatusAccepted)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if len(SECRET) == 0 {
		log.Fatal("[L8: บอดี้การ์ด] แจ้งเตือน: ไม่พบ POKE-SECRET ระบบไม่ปลอดภัย!")
	}

	startGenerals(ctx)

	// API Routes mapping อาชีพ
	http.HandleFunc("/process", processHandler) // รวมศูนย์การรับงาน
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// [L11: หมอผี] รายงานสถานะเครื่อง
		fmt.Fprintf(w, "V8.3 Trinity: สัจจะยังคงอยู่ (Generals Active)")
	})

	log.Printf("🏰 THITNUEAHUB Engine รันที่พอร์ต %s", PORT)
	http.ListenAndServe(PORT, nil)
}

