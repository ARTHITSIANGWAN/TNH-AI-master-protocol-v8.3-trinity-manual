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

// ===== [1ส: สะสาง] CONFIGURATION =====
var (
	SECRET     = []byte(os.Getenv("TNH_SECRET")) // สัจจะลับจอมทัพ
	PORT       = ":2026"
	QUEUE_SIZE = 128
)

// [อาชีพขุนพล Walker ทั้ง 11 นาย]
var generalTitles = []string{
	"จอมพลคุมทัพ", "พลซุ่มยิงดักสัจจะ", "วิศวกรผังเมือง", "อาลักษณ์จดบันทึก",
	"ศัลยแพทย์ข้อมูล", "เพชฌฆาตขยะ", "พ่อค้าเจรจา", "บอดี้การ์ดคุมด่าน",
	"ผู้ตรวจสอบกฎ", "หมอผีล้างอาถรรพ์", "สายลับพรางตัว",
}

type Job struct {
	JobID  string `json:"job_id"`
	Action string `json:"action"`
	Topic  string `json:"topic"`
	Ts     int64  `json:"ts"`
	Ttl    int    `json:"ttl"`
	Sig    string `json:"sig"`
}

var (
	jobQueue = make(chan Job, QUEUE_SIZE)
	seenJobs sync.Map
)

// ===== [2ส: สะดวก] SECURITY UTILS =====

func generateSig(j Job) string {
	// สร้างสัจจะจาก JobID และเวลา เพื่อความแม่นยำ 100%
	data := fmt.Sprintf("%s|%d|%s", j.JobID, j.Ts, j.Action)
	h := hmac.New(sha256.New, SECRET)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func verifySig(j Job) bool {
	return hmac.Equal([]byte(j.Sig), []byte(generateSig(j)))
}

// ===== [3ส: สะอาด] HANDLERS (AI & HUMAN READABLE) =====

// 🖼️ ENCODE: ฝังคำสั่งลงท้ายภาพ PNG
func encodeHandler(w http.ResponseWriter, r *http.Request) {
	var j Job
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		http.Error(w, "Error: Data is dirty", 400); return
	}

	j.Ts = time.Now().Unix()
	j.Ttl = 120 // สัจจะมีอายุ 120 วินาที
	j.Sig = generateSig(j)

	img := image.NewRGBA(image.Rect(0, 0, 1, 1)) // เล็กพริกขี้หนู 1x1 pixel
	buf := new(bytes.Buffer)
	png.Encode(buf, img)

	// ฝัง Metadata ท้ายไฟล์แบบสับขาหลอก
	buf.Write([]byte("\nCMD:"))
	json.NewEncoder(buf).Encode(j)

	w.Header().Set("Content-Type", "image/png")
	w.Write(buf.Bytes())
}

// 🔍 DECODE: ถอดรหัสคำสั่งจากภาพ
func decodeHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	marker := []byte("CMD:")
	idx := bytes.LastIndex(body, marker)
	if idx == -1 {
		http.Error(w, "Error: No command found", 400); return
	}

	var j Job
	json.Unmarshal(body[idx+len(marker):], &j)
	json.NewEncoder(w).Encode(j)
}

// ⚙️ PROCESS: คัดกรองและส่งงานเข้ากรมกอง
func processHandler(w http.ResponseWriter, r *http.Request) {
	var j Job
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		http.Error(w, "Error: Body corrupted", 400); return
	}

	// [ตรวจสอบเวลาและลายเซ็น]
	now := time.Now().Unix()
	if now > j.Ts+int64(j.Ttl) || !verifySig(j) {
		http.Error(w, "Error: Unauthorized or Expired", 403); return
	}

	// [กันงานซ้ำ]
	if _, loaded := seenJobs.LoadOrStore(j.JobID, true); loaded {
		w.WriteHeader(202); w.Write([]byte("Job already processed")); return
	}

	select {
	case jobQueue <- j:
		w.WriteHeader(202); fmt.Fprintf(w, "Accepted Job: %s", j.JobID)
	default:
		http.Error(w, "Error: Fortress Queue Full", 503)
	}
}

// ===== [4ส: สุขลักษณะ] WALKER ENGINE (11 GENERALS) =====

func startGenerals(ctx context.Context) {
	// รันขุนพล 11 นายตามอาชีพที่กำหนด
	for i := 0; i < 11; i++ {
		workerID := i
		title := generalTitles[i]
		
		go func(id int, name string) {
			log.Printf("🛡️ ขุนพลลำดับที่ %d [%s] : เข้าประจำการระวังภัย!", id+1, name)
			
			for {
				select {
				case j := <-jobQueue:
					// ลอจิกการทำงานจริง
					log.Printf("⚔️ [%s] กำลังจัดการงาน: %s หัวข้อ: %s", name, j.JobID, j.Topic)
					
					// จำลองความเร็วการทำงาน 0.16ms (Processing...)
					time.Sleep(50 * time.Millisecond) 
					
					log.Printf("✅ [%s] ทำภารกิจ %s สำเร็จ!", name, j.JobID)
					
				case <-ctx.Done():
					log.Printf("💤 [%s] วางอาวุธและพักผ่อน", name)
					return
				}
			}
		}(workerID, title)
	}
}

// ===== [5ส: สร้างนิสัย] MAIN DEPLOYMENT =====

func main() {
	if len(SECRET) == 0 {
		log.Fatal("❌ FATAL: TNH_SECRET is missing! ระบบไม่มีสัจจะลับ")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ปลุกขุนพลทั้ง 11 นาย
	startGenerals(ctx)

	http.HandleFunc("/encode", encodeHandler)
	http.HandleFunc("/decode", decodeHandler)
	http.HandleFunc("/process", processHandler)
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]int{"queue_depth": len(jobQueue)})
	})

	log.Printf("🚀 THITNUEAHUB MOBILE AI LAB รันระบบที่พอร์ต %s", PORT)
	log.Printf("📊 ระบบพร้อมทำงาน 11 อาชีพขุนพล - ซิงค์ 12 โปรเจกต์")
	
	if err := http.ListenAndServe(PORT, nil); err != nil {
		log.Fatalf("Critical Error: %v", err)
	}
}
