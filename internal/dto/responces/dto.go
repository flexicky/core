package responces

type ResponseWrapper struct {
	Ok      bool
	Message string
}

type LoginResponse struct {
	GrpcData    ResponseWrapper
	AccessToken *string
	ExpiresIn   *int64
}

type RegisterResponse struct {
	GrpcData      ResponseWrapper
	LoginResponse LoginResponse
}
