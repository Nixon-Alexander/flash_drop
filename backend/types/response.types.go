package types

type Response string

const (
	Success Response = "success"
	Error   Response = "error"
)

type ResponseType struct {
	Status  Response `json:"status"`
	Code    int      `json:"code"`
	Data    any      `json:"data"`
	Message string   `json:"message"`
}

type RepositoryErrorResponseType struct {
	HttpCode     int    `json:"http_code"`
	ErrorMessage string `json:"error_message"`
}
