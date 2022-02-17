package config

type Analytics struct {
	GatewayUrl      *string
	LogLevel        *string
	SigningKey      *string
	ProcessID       *string
	QuestionIndexes *[]int
	TargetValue     *int
}
