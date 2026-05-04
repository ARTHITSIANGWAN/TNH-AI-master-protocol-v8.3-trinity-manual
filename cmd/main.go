package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// [1ส: สะสาง] - Command Center Config
var (
	SECRET = []byte(os.Getenv("TNH_SECRET"))
	PORT   = ":2026"
	// คิวงานแยกตามประเภทอาชีพ (Priority Lanes)
	frontGateChan = make(chan Job, 2048) // L9, L8, L7
	coreLogicChan = make(chan Job, 2048) // L3, L4, L5
	outputChan    = make(chan Job, 1024) // L6, L10, L11
	seen          sync.Map               // [L5: เพชฌฆาตขยะ]
)

type Job struct {
	JobID   string      `json:"job_id"`
	Action  string      `json:"action"`
	Topic   string      `json:"topic"`
	Payload interface{} `json:"payload"`
	Ts      int64       `json:"ts"`
	Ttl     int         `json:"ttl"`
	Sig     string      `json:"sig"`
}

// ===== [3ส: สะอาด] THE HELMET (Core Logic) =====

func generateSovereignSig(j Job) string {
	// แก้บั๊ก rune: ใช้สัจจะจาก Sprintf เท่านั้น
	data := fmt.Sprintf("%s|%s|%s|%d", j.JobID, j.Action, j.Topic, j.Ts)
	h := hmac.New(sha256.New, SECRET)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func verifySovereignSig(j Job) bool {
	return hmac.Equal([]byte(j.Sig), []byte(generateSovereignSig(j)))
}

// ===== [5ส: สร้างนิสัย] THE 11 GENERALS (Deployment) =====

func startTheGenerals(ctx context.Context) {
	log.Println("🏰 [L1: จอมพล] ประกาศระดมพล 11 ขุนพลเข้าประจำการ!")

	// UNIT 1: กองหน้า (Security & Traffic) -> L9, L8, L7
	for i := 0; i < 3; i++ {
		go func() {
			for j := range frontGateChan {
				// [L9: พลซุ่มยิง] ดักสัจจะปลอม
				if !verifySovereignSig(j) {
					log.Printf("❌ [L9] สอยสัจจะปลอม: %s", j.JobID)
					continue
				}
				// [L7: วิศวกรผังเมือง] คุมการไหล
				coreLogicChan <- j
			}
		}()
	}

	// UNIT 2: กองกลาง (Logic & Data Surgery) -> L3, L4, L5, L2
	for i := 0; i < 5; i++ {
		go func() {
			for j := range coreLogicChan {
				// [L2: อาลักษณ์] บันทึกลง Blackbox (Log)
				log.Printf("📝 [L2] จดบันทึกสัจจะงาน: %s", j.JobID)
				
				// [L4: นักสืบ] วิเคราะห์ Payload
				// [L3: ศัลยแพทย์] จัดรูปขบวนข้อมูล
				outputChan <- j
			}
		}()
	}

	// UNIT 3: กองหลัง (Business & Healing) -> L6, L10, L11
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		for {
			select {
			case j := <-outputChan:
				// [L6: พ่อค้า] จัดการ SME Logic
				// [L10: ผู้ตรวจสอบ] เช็ก Compliance
				log.Printf("✅ [L6/L10] ภารกิจสำเร็จ: %s", j.JobID)
			case <-ticker.C:
				// [L11: หมอผี] ทำพิธีล้าง Memory Leak
				seen = sync.Map{}
				log.Println("🔮 [L11] ล้างอาถรรพ์ข้อมูลเก่าเรียบร้อย (Self-Healing)")
			case <-ctx.Done():
				return
			}
		}
	}()
}

// ===== [4ส: สุขลักษณะ] THE DARK RELAY (Handlers) =====

func encodeHandler(w http.ResponseWriter, r *http.Request) {
	var j Job
	json.NewDecoder(r.Body).Decode(&j)
	j.Ts, j.Ttl = time.Now().Unix(), 180
	j.Sig = generateSovereignSig(j)

	// [L3: ศัลยแพทย์] สร้างภาพพรางสัจจะ
	img := image.NewRGBA(image.Rect(0, 0, 1, 1)) // Zero-Garbage: 1x1 pixel
	buf := new(bytes.Buffer)
	png.Encode(buf, img)
	buf.Write([]byte("\nCMD:"))
	json.NewEncoder(buf).Encode(j)

	w.Header().Set("Content-Type", "image/png")
	w.Write(buf.Bytes())
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	var j Job
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		http.Error(w, "สาส์นสกปรก", 400); return
	}

	// [L5: เพชฌฆาตขยะ] กันงานซ้ำ
	if _, loaded := seen.LoadOrStore(j.JobID, true); loaded {
		w.WriteHeader(202); w.Write([]byte("สัจจะซ้ำ")); return
	}

	select {
	case frontGateChan <- j:
		w.WriteHeader(http.StatusAccepted)
	default:
		http.Error(w, "กองทัพรับงานไม่ไหว", 503)
	}
}

func main() {
	if len(SECRET) == 0 {
		log.Fatal("❌ [L8: บอดี้การ์ด] แจ้งเตือน: ไม่พบ SECRET! ระบบเปิดเผยสัจจะเกินไป!")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startTheGenerals(ctx)

	http.HandleFunc("/encode", encodeHandler)   // L3: ศัลยแพทย์
	http.HandleFunc("/process", processHandler) // รวมศูนย์
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("V8.3 Trinity: สัจจะยังคงอยู่ 🏰"))
	})

	log.Printf("🚀 THITNUEAHUB Engine รันที่พอร์ต %s [0.16ms MODE]", PORT)
	http.ListenAndServe(PORT, nil)
}
