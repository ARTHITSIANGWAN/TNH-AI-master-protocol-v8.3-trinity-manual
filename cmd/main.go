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

// [1ส: สะสาง] - แยกแยะขุนพลและอาวุธ
var (
	SECRET     = []byte(getEnv("TNH_SECRET", "THITNUEAHUB_CORE_2026"))
	PORT       = ":2026"
	QUEUE_SIZE = 128
	WORKER_QTY = 4 // จำนวนขุนพลรบหลัก
)

// [2ส: สะดวก] - วางรูปแบบ Job ให้คมชัด
type Job struct {
	JobID  string `json:"job_id"`
	Action string `json:"action"`
	Topic  string `json:"topic"`
	Ts     int64  `json:"ts"`  // เวลาสร้างสัจจะ
	Ttl    int    `json:"ttl"` // อายุของสาส์น
	Sig    string `json:"sig"` // ลายเซ็นจอมทัพ
}

var (
	jobQueue = make(chan Job, QUEUE_SIZE)
	seenJobs sync.Map // [5ส: สร้างนิสัย] - กันงานซ้ำ (Idempotency)
)

// ===== [3ส: สะอาด] - ขุนพลฝ่ายตรวจสอบ (Security Layer) =====

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" { return v }
	return d
}

// สร้างลายเซ็นจอมทัพ
func generateSignature(j Job) string {
	// สร้างสัจจะจากข้อมูลหลัก (JobID + Ts + Action)
	data := fmt.Sprintf("%s|%d|%s", j.JobID, j.Ts, j.Action)
	h := hmac.New(sha256.New, SECRET)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func verifySovereignty(j Job) bool {
	return hmac.Equal([]byte(j.Sig), []byte(generateSignature(j)))
}

// ===== [4ส: สุขลักษณะ] - ระบบไหลเวียนข้อมูล (Core Handlers) =====

// 🖼️ Encode: ฝังคำสั่งลงท้ายไฟล์ภาพ (Steganography)
func encodeHandler(w http.ResponseWriter, r *http.Request) {
	var j Job
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		http.Error(w, "สาส์นสกปรก", 400); return
	}

	j.Ts = time.Now().Unix()
	j.Ttl = 120 // สัจจะมีอายุ 2 นาที
	j.Sig = generateSignature(j)

	// สร้างภาพเปล่า (Canvas) 300x300
	img := image.NewRGBA(image.Rect(0, 0, 300, 300))
	buf := new(bytes.Buffer)
	png.Encode(buf, img)

	// [สับขาหลอก] แนบสัจจะไว้ท้ายภาพ
	buf.Write([]byte("\nCMD:"))
	json.NewEncoder(buf).Encode(j)

	w.Header().Set("Content-Type", "image/png")
	w.Write(buf.Bytes())
}

// 🔍 Decode: อ่านสัจจะจากภาพ
func decodeHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	marker := []byte("CMD:")
	idx := bytes.LastIndex(body, marker)
	if idx == -1 {
		http.Error(w, "ไม่พบสัจจะในภาพ", 400); return
	}

	var j Job
	if err := json.Unmarshal(body[idx+len(marker):], &j); err != nil {
		http.Error(w, "สาส์นบิดเบือน", 400); return
	}
	json.NewEncoder(w).Encode(j)
}

// ⚙️ Process: รับงานเข้าสู่คิวขุนพล
func processHandler(w http.ResponseWriter, r *http.Request) {
	var j Job
	json.NewDecoder(r.Body).Decode(&j)

	// [ขุนพลคุมเวลา] - Time Gate
	now := time.Now().Unix()
	if now > j.Ts+int64(j.Ttl) {
		http.Error(w, "สาส์นหมดอายุ (TTL Expired)", 403); return
	}

	// [ขุนพลตรวจสอบ] - Signature Check
	if !verifySovereignty(j) {
		http.Error(w, "สัจจะปลอม (Invalid Signature)", 401); return
	}

	// [ขุนพลเพชฌฆาตขยะ] - Duplicate Check
	if _, loaded := seenJobs.LoadOrStore(j.JobID, true); loaded {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("งานนี้ทำแล้ว (Duplicate)"))
		return
	}

	select {
	case jobQueue <- j:
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "รับสาส์น %s เข้ากรมกอง", j.JobID)
	default:
		http.Error(w, "คิวเต็ม (Overload)", 503)
	}
}

// 📅 Scheduler: ตั้งเวลาทำงานล่วงหน้า
func scheduleHandler(w http.ResponseWriter, r *http.Request) {
	var j Job
	json.NewDecoder(r.Body).Decode(&j)
	
	go func() {
		delay := time.Until(time.Unix(j.Ts, 0))
		if delay > 0 { time.Sleep(delay) }

		// ยิงเข้า Process ตัวเอง
		b, _ := json.Marshal(j)
		http.Post("http://localhost"+PORT+"/process", "application/json", bytes.NewReader(b))
	}()
	w.Write([]byte("ตั้งเวลาสัจจะเรียบร้อย"))
}

// ===== [DEPLOYMENT] - ขุนพลประจำการ =====

func worker(ctx context.Context, id int) {
	log.Printf("🛡️ ขุนพลที่ %d พร้อมรบ!", id)
	for {
		select {
		case j := <-jobQueue:
			log.Printf("[Worker %d] ปฏิบัติงาน: %s ในหัวข้อ %s", id, j.Action, j.Topic)
			// จำลองการทำงานหนัก (0.16ms ในอุดมคติ)
			time.Sleep(100 * time.Millisecond)
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ปล่อยขุนพลรบหลัก 4 นาย (Worker Pool)
	for i := 1; i <= WORKER_QTY; i++ {
		go worker(ctx, i)
	}

	http.HandleFunc("/encode", encodeHandler)     // ฝัง
	http.HandleFunc("/decode", decodeHandler)     // อ่าน
	http.HandleFunc("/process", processHandler)   // รัน
	http.HandleFunc("/schedule", scheduleHandler) // ตั้งเวลา
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ความยาวคิวปัจจุบัน: %d", len(jobQueue))
	})

	log.Printf("🚀 THITNUEAHUB Sovereign Engine รันบนพอร์ต %s", PORT)
	if err := http.ListenAndServe(PORT, nil); err != nil {
		log.Fatalf("ระบบล่ม: %v", err)
	}
}
