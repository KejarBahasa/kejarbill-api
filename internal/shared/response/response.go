package response

import "github.com/gofiber/fiber/v3"

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

type Meta struct {
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

type Response[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Errors  any    `json:"errors,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// Success send response 200 OK by default.
// data send as is, if nil will show "data": null.
func Success[T any](c fiber.Ctx, message string, data T, statusCode ...int) error {
	code := fiber.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}

	return c.Status(code).JSON(Response[T]{
		Status:  StatusSuccess,
		Message: message,
		Data:    data,
	})
}

// SuccessWithMeta for response that need pagination
func SuccessWithMeta[T any](c fiber.Ctx, message string, data T, meta *Meta, statusCode ...int) error {
	code := fiber.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}

	return c.Status(code).JSON(Response[T]{
		Status:  StatusSuccess,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Error for handle all type of error response
func Error(c fiber.Ctx, statusCode int, message string, errors any) error {
	return c.Status(statusCode).JSON(Response[any]{
		Status:  StatusError,
		Message: message,
		Data:    nil,
		Errors:  errors,
	})
}
