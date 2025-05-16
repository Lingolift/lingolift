package params

// AssessmentRequest 评测请求参数
type AssessmentRequest struct {
	RefText          string  `json:"ref_text" form:"ref_text" query:"ref_text" validate:"required"`
	ServerEngineType string  `json:"server_engine_type" form:"server_engine_type" query:"server_engine_type" default:"16k_en"`
	ScoreCoeff       float64 `json:"score_coeff" form:"score_coeff" query:"score_coeff" default:"1.1"`
	EvalMode         int64   `json:"eval_mode" form:"eval_mode" query:"eval_mode" default:"0"`
	TextMode         int64   `json:"text_mode" form:"text_mode" query:"text_mode" default:"0"`
}
