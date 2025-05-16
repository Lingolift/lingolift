package response

import (
	"github.com/tencentcloud/tencentcloud-speech-sdk-go/soe"
)

// 评测结果响应
type AssessmentResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Result  *soe.SentenceInfo `json:"result,omitempty"`
}
