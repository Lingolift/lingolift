package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"lingolift/config"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/tencentcloud/tencentcloud-speech-sdk-go/common"
	"github.com/tencentcloud/tencentcloud-speech-sdk-go/soe"
)

var (
	// WebSocket 升级器
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源
		},
	}
)

// 评测请求参数
type AssessmentRequest struct {
	RefText          string  `json:"ref_text" validate:"required"`
	ServerEngineType string  `json:"server_engine_type" default:"16k_en"`
	ScoreCoeff       float64 `json:"score_coeff" default:"1.1"`
	EvalMode         int64   `json:"eval_mode" default:"0"`
	TextMode         int64   `json:"text_mode" default:"0"`
}

// 评测结果响应
type AssessmentResponse struct {
	Status string     `json:"status"`
	Result *SOEResult `json:"result,omitempty"`
	Error  string     `json:"error,omitempty"`
}

// 自定义语音评测结果结构
type SOEResult struct {
	OverallScore   float64       `json:"overall_score,omitempty"`
	Words          []soe.WordRsp `json:"words,omitempty"`
	PronAccuracy   float64       `json:"pron_accuracy,omitempty"`
	PronFluency    float64       `json:"pron_fluency,omitempty"`
	PronCompletion float64       `json:"pron_completion,omitempty"`
}

// 流式评测监听器
type StreamListener struct {
	Conn       *websocket.Conn
	ResultChan chan *SOEResult
	ErrorChan  chan error
	Complete   chan struct{}
}

func NewStreamListener(conn *websocket.Conn) *StreamListener {
	return &StreamListener{
		Conn:       conn,
		ResultChan: make(chan *SOEResult, 10),
		ErrorChan:  make(chan error, 1),
		Complete:   make(chan struct{}),
	}
}

func (l *StreamListener) OnRecognitionStart(response *soe.SpeakingAssessmentResponse) {
	log.Printf("OnRecognitionStart: %s", response.VoiceID)
	l.sendResponse("start", nil, nil)
}

func (l *StreamListener) OnIntermediateResults(response *soe.SpeakingAssessmentResponse) {
	log.Printf("OnIntermediateResults: %+v", response.Result)

	if len(response.Result.Words) > 0 {
		result := &SOEResult{
			OverallScore:   response.Result.SuggestedScore,
			Words:          response.Result.Words,
			PronAccuracy:   response.Result.PronAccuracy,
			PronFluency:    response.Result.PronFluency,
			PronCompletion: response.Result.PronCompletion,
		}
		l.ResultChan <- result
		l.sendResponse("intermediate", result, nil)
	}
}

func (l *StreamListener) OnRecognitionComplete(response *soe.SpeakingAssessmentResponse) {
	log.Printf("OnRecognitionComplete: %+v", response.Result)

	if len(response.Result.Words) > 0 {
		result := &SOEResult{
			OverallScore:   response.Result.SuggestedScore,
			Words:          response.Result.Words,
			PronAccuracy:   response.Result.PronAccuracy,
			PronFluency:    response.Result.PronFluency,
			PronCompletion: response.Result.PronCompletion,
		}
		l.ResultChan <- result
		l.sendResponse("complete", result, nil)
	}

	close(l.Complete)
}

func (l *StreamListener) OnFail(response *soe.SpeakingAssessmentResponse, err error) {
	log.Printf("OnFail: %v", err)
	l.ErrorChan <- err
	l.sendResponse("error", nil, err)
	close(l.Complete)
}

func (l *StreamListener) sendResponse(status string, result *SOEResult, err error) {
	log.Println("1")
	// 添加连接状态检查
	if l.Conn == nil {
		log.Println("WebSocket connection is nil")
		return
	}
	log.Println("2")
	// 使用同步锁防止并发写入
	l.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	defer l.Conn.SetWriteDeadline(time.Time{})
	log.Println("3")
	response := AssessmentResponse{
		Status: status,
		Result: result,
	}
	log.Println("4")
	if err != nil {
		response.Error = err.Error()
	}
	log.Println("5")
	if err := l.Conn.WriteJSON(response); err != nil {
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			log.Println("WebSocket closed normally")
		} else if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Printf("Unexpected WebSocket close: %v", err)
		} else {
			log.Printf("WebSocket write error: %v", err)
		}
	}

	log.Println("6")
}

// 定义结束标志消息类型
type EndMessage struct {
	Type string `json:"type"`
}

// WebSocket处理函数
func HandleWebSocket(c echo.Context) error {
	// 升级HTTP连接为WebSocket连接
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return err
	}
	defer conn.Close()

	// 读取初始配置消息
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Read initial message error: %v", err)
		return nil
	}

	// 解析配置
	var req AssessmentRequest
	if err := json.Unmarshal(message, &req); err != nil {
		log.Printf("Parse config error: %v", err)
		conn.WriteJSON(AssessmentResponse{
			Status: "error",
			Error:  "Invalid configuration",
		})
		return nil
	}

	// 创建流式监听器
	listener := NewStreamListener(conn)

	// 认证信息
	credential := common.NewCredential(config.G.Speech.SecretID, config.G.Speech.SecretKey)

	fmt.Println(req.RefText)
	// 创建识别器
	recognizer := soe.NewSpeechRecognizer(config.G.Speech.AppID, credential, listener)
	recognizer.VoiceFormat = soe.AudioFormatPCM
	recognizer.RefText = req.RefText
	recognizer.ServerEngineType = req.ServerEngineType // 使用客户端选择的引擎类型
	recognizer.ScoreCoeff = 1.1
	// recognizer.SentenceInfoEnabled = 1
	recognizer.EvalMode = req.EvalMode
	recognizer.TextMode = req.TextMode

	// 启动识别器
	if err := recognizer.Start(); err != nil {
		log.Printf("Recognizer start error: %v", err)
		conn.WriteJSON(AssessmentResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return nil
	}

	// 确保识别器在结束时停止
	defer recognizer.Stop()

	// 处理WebSocket消息
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer recognizer.Stop() // 确保识别器停止

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				listener.ErrorChan <- err
				return // 直接返回不再重试
			}

			// 处理结束标志
			if messageType == websocket.TextMessage {
				var endMsg EndMessage
				if json.Unmarshal(message, &endMsg) == nil && endMsg.Type == "end" {
					log.Println("收到客户端结束标志")
					return // 退出循环，让defer处理识别器停止
				}
				// 忽略其他文本消息
				continue
			}

			// 只处理二进制消息
			if messageType != websocket.BinaryMessage {
				continue
			}

			// 发送音频数据到识别器
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
		log.Println("Assessment completed")
		wg.Wait()
		return nil
	}
}

// WebSocket处理函数
func HandleWebSocketV2(c echo.Context) error {
	// 升级HTTP连接为WebSocket连接
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return err
	}
	defer conn.Close()

	// 读取初始配置消息
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Read initial message error: %v", err)
		return nil
	}

	// 解析配置
	var req AssessmentRequest
	if err := json.Unmarshal(message, &req); err != nil {
		log.Printf("Parse config error: %v", err)
		conn.WriteJSON(AssessmentResponse{
			Status: "error",
			Error:  "Invalid configuration",
		})
		return nil
	}

	// 创建流式监听器
	listener := NewStreamListener(conn)

	// 认证信息
	credential := common.NewCredential(config.G.Speech.SecretID, config.G.Speech.SecretKey)

	fmt.Println(req.RefText)
	// 创建识别器
	recognizer := soe.NewSpeechRecognizer(config.G.Speech.AppID, credential, listener)
	recognizer.VoiceFormat = soe.AudioFormatPCM
	recognizer.RefText = req.RefText
	recognizer.ServerEngineType = "16k_en"
	recognizer.ScoreCoeff = 1.1
	// recognizer.SentenceInfoEnabled = 1
	recognizer.EvalMode = req.EvalMode
	recognizer.TextMode = req.TextMode

	// 启动识别器
	if err := recognizer.Start(); err != nil {
		log.Printf("Recognizer start error: %v", err)
		conn.WriteJSON(AssessmentResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return nil
	}

	// 确保识别器在结束时停止
	defer recognizer.Stop()

	// 处理WebSocket消息
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer recognizer.Stop() // 确保识别器停止

		for {
			select {
			case <-listener.Complete:
				return
			default:
				// 读取音频数据
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("WebSocket read error: %v", err)
					}
					listener.ErrorChan <- err
					return // 直接返回不再重试
				}

				// 忽略非二进制消息
				if messageType != websocket.BinaryMessage {
					continue
				}

				// 处理结束标志
				if messageType == websocket.TextMessage {
					var endMsg EndMessage
					if json.Unmarshal(message, &endMsg) == nil && endMsg.Type == "end" {
						log.Println("收到客户端结束标志")
						// 不需要在这里调用recognizer.Stop()，让客户端处理
						return
					}
					// 忽略其他文本消息
					continue
				}

				// 检查并补全音频数据长度为偶数
				if len(message)%2 != 0 {
					log.Printf("音频数据长度为奇数，补全前长度: %d", len(message))
					paddedMessage := make([]byte, len(message)+1)
					copy(paddedMessage, message)
					paddedMessage[len(message)] = 0 // 补零
					message = paddedMessage
					log.Printf("补全后长度: %d", len(message))
				}

				fmt.Println(3333)
				// 发送音频数据到识别器
				if err := recognizer.Write(message); err != nil {
					log.Printf("Recognizer write error: %v", err)
					listener.ErrorChan <- err
					return // 直接返回不再重试
				}
			}
		}
	}()

	// 监听结果
	select {
	case err := <-listener.ErrorChan:
		log.Printf("Assessment error: %v", err)
		return nil

	case <-listener.Complete:
		log.Println("Assessment completed")
		wg.Wait()
		return nil
	}
}
