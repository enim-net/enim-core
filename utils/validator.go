package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/enim-net/enim-core/response"
)

func ValidateUUID(c *fiber.Ctx, key string) (uuid.UUID, error) {
	_uuid, err := uuid.Parse(c.Params(key))
	if err != nil {
		return uuid.Nil, response.ErrorHandler(c, err)
	}

	return _uuid, nil
}
