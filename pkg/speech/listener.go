package speech

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tencentcloud/tencentcloud-speech-sdk-go/soe"
)

// StreamListener
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
	log.Printf("OnIntermediateResults: 整体得分=%.2f, 准确率=%.2f, 流畅度=%.2f",
		response.Result.SuggestedScore,
		response.Result.PronAccuracy,
		response.Result.PronFluency)

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
	log.Printf("语音识别评测结果完成: 整体得分=%.2f, 准确率=%.2f, 流畅度=%.2f",
		response.Result.SuggestedScore,
		response.Result.PronAccuracy,
		response.Result.PronFluency)

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
	log.Println("准备发送响应:", status)
	if l.Conn == nil {
		log.Println("WebSocket connection is nil")
		return
	}

	// 使用写锁防止并发写入
	l.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	defer l.Conn.SetWriteDeadline(time.Time{})

	response := AssessmentResponse{
		Status: status,
		Result: result,
	}
	if err != nil {
		response.Error = err.Error()
	}

	if err := l.Conn.WriteJSON(response); err != nil {
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			log.Println("WebSocket closed normally")
		} else if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Printf("Unexpected WebSocket close: %v", err)
		} else {
			log.Printf("WebSocket write error: %v", err)
		}
	}
}

// AssessmentRequest 评测请求参数
type AssessmentRequest struct {
	RefText          string  `json:"ref_text" validate:"required"`
	ServerEngineType string  `json:"server_engine_type" default:"16k_en"`
	ScoreCoeff       float64 `json:"score_coeff" default:"1.1"`
	EvalMode         int64   `json:"eval_mode" default:"0"`
	TextMode         int64   `json:"text_mode" default:"0"`
	IsSaveAudioFile  bool    `json:"is_save_audio_file" default:"false"`
}

func (req *AssessmentRequest) Validator() error {
	// 可以添加其他参数的验证逻辑
	return nil
}

type AssessmentResponse struct {
	Status string     `json:"status"`
	Result *SOEResult `json:"result,omitempty"`
	Error  string     `json:"error,omitempty"`
}

type SOEResult struct {
	OverallScore   float64       `json:"overall_score,omitempty"`
	Words          []soe.WordRsp `json:"words,omitempty"`
	PronAccuracy   float64       `json:"pron_accuracy,omitempty"`
	PronFluency    float64       `json:"pron_fluency,omitempty"`
	PronCompletion float64       `json:"pron_completion,omitempty"`
}
