package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// [1. ส: สะสาง] - ระบบควบคุมส่วนกลาง (Command Center)
var (
	SECRET     = []byte(getenv("TNH_SECRET", "THITNUEAHUB_SOVEREIGN_2026"))
	PORT       = ":2026"
	QUEUE_SIZE = 1024
	WORKERS    = 11 // ประจำการตามจำนวน 11 ขุนพล
)

// [2. ส: สะดวก] - โครงสร้างสัจจะดิจิทัล (Job Structure)
type Job struct {
	JobID  string `json:"job_id"`
	Action string `json:"action"`
	Topic  string `json:"topic"`
	Ts     int64  `json:"ts"`
	Ttl    int    `json:"ttl"`
	Sig    string `json:"sig"`
}

var (
	queue = make(chan Job, QUEUE_SIZE)
	seen  sync.Map // สมองส่วนความจำกันงานซ้ำ
)

// [3. ส: สะอาด] - ฟังก์ชันคิดวิเคราะห์แยกแยะ (Security Helpers)
func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" { return v }
	return d
}

func generateSignature(j Job) string {
	data := j.JobID + j.Action + j.Topic + string(rune(j.Ts))
	h := hmac.New(sha256.New, SECRET)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func verifyJob(j Job) bool {
	return hmac.Equal([]byte(j.Sig), []byte(generateSignature(j)))
}

// [4. ส: สุขลักษณะ] - ส่วนติดต่อสื่อสารระหว่างขุนพล (Handlers)

// L3 Artist (น้ำอิง): ฝังคำสั่งสัจจะลงภาพ (PNG Encoding)
func encodeHandler(w http.ResponseWriter, r *http.Request) {
	var j Job
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		http.Error(w, "Invalid Matrix", 400)
		return
	}
	j.Ts, j.Ttl = time.Now().Unix(), 180
	j.Sig = generateSignature(j)

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	buf := new(bytes.Buffer)
	png.Encode(buf, img)
	payload, _ := json.Marshal(j)
	buf.Write([]byte("\nCMD:")) 
	buf.Write(payload)

	w.Header().Set("Content-Type", "image/png")
	w.Write(buf.Bytes())
}

// L5 Auditor (ไอ้จ๊อด): ตรวจสอบความสะอาดและเข้าคิว
func processHandler(w http.ResponseWriter, r *http.Request) {
	var j Job
	json.NewDecoder(r.Body).Decode(&j)

	if time.Now().Unix() > j.Ts+int64(j.Ttl) || !verifyJob(j) {
		http.Error(w, "Unauthorized or Expired", 401)
		return
	}

	if _, loaded := seen.LoadOrStore(j.JobID, true); loaded {
		w.Write([]byte("สัจจะนี้ได้รับการปฏิบัติแล้ว"))
		return
	}

	select {
	case queue <- j:
		w.WriteHeader(202)
	default:
		http.Error(w, "คิวเต็ม", 503)
	}
}

// [5. ส: สร้างนิสัย] - 11 ขุนพลปฏิบัติการ (Worker Pool)
func worker(ctx context.Context, id int) {
	log.Printf("[L%d] ขุนพลพร้อมประจำการ", id+1)
	for {
		select {
		case j := <-queue:
			// L11 Balancer: ระบบวิเคราะห์การกระจายงาน
			if j.Topic == "edge" {
				log.Printf("[ขุนพล %d] ส่งต่องานไปสมรภูมิ Edge (V5 Curator)", id+1)
			} else {
				log.Printf("[ขุนพล %d] ปฏิบัติภารกิจ: %s บนหัวข้อ %s", id+1, j.Action, j.Topic)
			}
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < WORKERS; i++ {
		go worker(ctx, i)
	}

	http.HandleFunc("/encode", encodeHandler)
	http.HandleFunc("/process", processHandler)
	
	log.Println("🏰 TNH-AI V8.3: 11 Generals Sovereign Engine พร้อมรบที่พอร์ต", PORT)
	http.ListenAndServe(PORT, nil)
}

