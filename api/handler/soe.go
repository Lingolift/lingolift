package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"lingolift/config"
	"lingolift/pkg/speech"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/tencentcloud/tencentcloud-speech-sdk-go/common"
	"github.com/tencentcloud/tencentcloud-speech-sdk-go/soe"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type EndMessage struct {
	Type string `json:"type"`
}

// StreamAssessment
func StreamAssessment(c echo.Context) error {
	// 升级HTTP连接为WebSocket连接
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return err
	}
	defer conn.Close()

	log.Println("新的WebSocket连接已建立")

	mimeType := c.Request().Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "audio/webm;codecs=opus"
	}
	log.Printf("等待接受的音频类型: %s", mimeType)

	// 读取初始配置消息
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Read initial message error: %v", err)
		return nil
	}

	// 解析配置
	var req speech.AssessmentRequest
	if err = json.Unmarshal(message, &req); err != nil {
		log.Printf("Parse config error: %v", err)
		conn.WriteJSON(speech.AssessmentResponse{
			Status: "error",
			Error:  "Invalid configuration",
		})
		return nil
	}

	if req.ScoreCoeff <= 0 {
		req.ScoreCoeff = 1
	}

	if isSentence(req.RefText) {
		req.EvalMode = 1
	}

	log.Printf("收到配置: RefText=%s, EngineType=%s, EvalMode=%d, ScoreCoeff=%.2f",
		req.RefText, req.ServerEngineType, req.EvalMode, req.ScoreCoeff)

	// 创建流式监听器
	listener := speech.NewStreamListener(conn)

	// 认证信息
	credential := common.NewCredential(config.G.Speech.SecretID, config.G.Speech.SecretKey)

	// 创建识别器
	recognizer := soe.NewSpeechRecognizer(config.G.Speech.AppID, credential, listener)
	recognizer.VoiceFormat = soe.AudioFormatWav
	recognizer.RefText = req.RefText
	recognizer.ServerEngineType = req.ServerEngineType
	recognizer.ScoreCoeff = req.ScoreCoeff // 使用客户端传递的系数
	recognizer.EvalMode = req.EvalMode
	recognizer.TextMode = req.TextMode

	// 启动识别器
	log.Println("准备启动识别器...")
	if err = recognizer.Start(); err != nil {
		log.Printf("Recognizer start error: %v", err)
		conn.WriteJSON(speech.AssessmentResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return nil
	}
	log.Println("识别器已成功启动")

	// 确保识别器在结束时停止
	defer func() {
		log.Println("正在停止识别器...")
		recognizer.Stop()
		log.Println("识别器已停止")
	}()

	// 创建音频文件用于调试
	var audioFile *os.File
	if req.IsSaveAudioFile {
		fileName := generateUniqueFilename(mimeType)
		audioFile, err := os.Create(fileName)
		if err != nil {
			log.Printf("创建音频文件失败: %v", err)
		} else {
			defer audioFile.Close()
			log.Printf("音频将保存到: %s", fileName)
		}
	}

	var (
		totalBytes  int
		startTime   = time.Now()
		audioChunks [][]byte
	)

	// 处理WebSocket消息
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer log.Println("音频处理协程已退出")

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				listener.ErrorChan <- err
				return
			}

			// 处理结束标志
			if messageType == websocket.TextMessage {
				var endMsg EndMessage
				if json.Unmarshal(message, &endMsg) == nil && endMsg.Type == "end" {
					log.Println("收到客户端结束标志")

					// 计算音频时长
					duration := float64(totalBytes) / (16000 * 2) // 16kHz, 16bit, 单声道
					log.Printf("音频接收完成: Total=%dByte, Estimated duration=%.2fs, cost=%.2fs",
						totalBytes, duration, time.Since(startTime).Seconds())

					// 主动通知SDK音频传输结束
					log.Println("通知识别器音频传输结束")
					recognizer.Stop()

					return
				}
				// 忽略其他文本消息
				continue
			}

			// 只处理二进制消息
			if messageType != websocket.BinaryMessage {
				continue
			}

			// 记录音频数据
			totalBytes += len(message)
			audioChunks = append(audioChunks, message)

			if req.IsSaveAudioFile {
				// 写入文件（用于调试）
				if audioFile != nil {
					if _, err := audioFile.Write(message); err != nil {
						log.Printf("写入音频文件失败: %v", err)
					}
				}
			}

			// 发送音频数据到识别器
			log.Printf("发送音频块到识别器: Size=%dByte, Total=%dByte", len(message), totalBytes)
			if err := recognizer.Write(message); err != nil {
				log.Printf("Recognizer write error: %v", err)
				listener.ErrorChan <- err
				return
			}
		}
	}()

	// 监听结果
	select {
	case err := <-listener.ErrorChan:
		log.Printf("Assessment error: %v", err)
		return nil

	case <-listener.Complete:
		log.Println("识别器流程已完整结束: Assessment completed")
		wg.Wait()
		return nil
	}
}

// 生成唯一文件名
func generateUniqueFilename(mimeType string) string {
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("audio_%s.wav", timestamp)
}

func isSentence(s string) bool {
	// 检查是否包含句子分隔符
	if strings.ContainsAny(s, ".?!") {
		return true
	}

	// 检查是否包含空格
	if strings.Contains(s, " ") {
		return true
	}

	// 检查是否包含多个单词（通过判断是否有大写字母在非首字符位置）
	runes := []rune(s)
	if len(runes) > 1 {
		for i := 1; i < len(runes); i++ {
			if unicode.IsUpper(runes[i]) {
				return true
			}
		}
	}

	return false
}
