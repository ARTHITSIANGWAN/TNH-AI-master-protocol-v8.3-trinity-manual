/**
 * 🛡️ TNH-ZERO-TRUST-API-CONNECTOR
 * วัตถุประสงค์: ดึงค่า App Confidence Score แบบ Real-time จาก Cloudflare
 */

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// โครงสร้างข้อมูลสำหรับรับค่าจาก Cloudflare API
type CloudflareAppResponse struct {
	Result []struct {
		Name            string  `json:"name"`
		ConfidenceScore float64 `json:"app_confidence_score"`
	} `json:"result"`
}

const (
	CF_API_URL  = "https://api.cloudflare.com/client/v4/accounts/%s/zero_trust/devices/applications"
	ACCOUNT_ID  = "YOUR_CLOUDFLARE_ACCOUNT_ID" // 🆔 ใส่ Account ID ของบอส
	API_TOKEN   = "YOUR_API_TOKEN"              // 🔑 ใส่ API Token ของบอส
)

// ฟังก์ชันดึงคะแนนแบบ Real-time
func fetchConfidenceScore() {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf(CF_API_URL, ACCOUNT_ID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+API_TOKEN)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ การเชื่อมต่อล้มเหลว: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var data CloudflareAppResponse
	json.Unmarshal(body, &data)

	// แสดงผลคะแนนของขุนพลแต่ละนาย
	for _, app := range data.Result {
		status := "✅ ผ่าน"
		if app.ConfidenceScore < 3.0 {
			status = "❌ เสี่ยง"
		}
		fmt.Printf("🛡️ แอป: %-15s | คะแนน: %.2f | สถานะ: %s\n", app.Name, app.ConfidenceScore, status)
	}
}

func main() {
	fmt.Println("🚀 กำลังตรวจสอบสัจจะข้อมูลจาก Cloudflare Zero Trust...")
	fetchConfidenceScore()
}
