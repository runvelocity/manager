package handlers

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/runvelocity/manager/models"
	"gorm.io/gorm"
)

// deleteFunctionHandler is the HTTP handler for DELETE /functions/{name}.
func DeleteFunctionHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	uuid := c.Param("uuid")
	var function models.Function
	result := db.Where("uuid = ?", uuid).First(&function)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.String(http.StatusNotFound, "Function does not exist")
		} else {
			errObj := models.ErrorResponse{Message: result.Error.Error()}
			return c.JSON(http.StatusInternalServerError, errObj)
		}
	}
	result = db.Where("uuid = ?", uuid).Delete(&function)
	if result.Error != nil {
		errObj := models.ErrorResponse{Message: result.Error.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	return c.JSON(http.StatusOK, function)
}
