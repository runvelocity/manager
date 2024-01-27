package handlers

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/runvelocity/manager/models"
	"gorm.io/gorm"
)

// getFunctions is the HTTP handler for GET /functions.
func GetFunctionsHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	var functions []models.Function
	result := db.Find(&functions)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			errObj := models.ErrorResponse{Message: "Function does not exist"}
			return c.JSON(http.StatusBadRequest, errObj)
		} else {
			errObj := models.ErrorResponse{Message: result.Error.Error()}
			return c.JSON(http.StatusInternalServerError, errObj)
		}
	}

	functionsResponse := models.FunctionsResponse{Functions: functions}
	return c.JSON(http.StatusOK, functionsResponse)
}

// getFunction is the HTTP handler for GET /functions/{uuid}.
func GetFunctionHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	uuid := c.Param("uuid")
	var function models.Function

	result := db.Where("uuid = ?", uuid).First(&function)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			errObj := models.ErrorResponse{Message: "Function does not exist"}
			return c.JSON(http.StatusBadRequest, errObj)
		} else {
			errObj := models.ErrorResponse{Message: result.Error.Error()}
			return c.JSON(http.StatusInternalServerError, errObj)
		}
	}

	return c.JSON(http.StatusOK, function)
}
