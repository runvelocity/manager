package handlers

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/runvelocity/manager/models"
	"gorm.io/gorm"
)

func UpdateFunctionHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	var function models.Function
	err := c.Bind(&function)
	if err != nil {
		errObj := models.ErrorResponse{Message: "Error while reading request body"}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	uuid := c.Param("uuid")
	var resObj models.Function
	result := db.Where("uuid = ?", uuid).First(&resObj)

	if result.Error == nil {
		db.Where("uuid = ?", uuid).Model(&resObj).Updates(function)
		return c.String(http.StatusOK, "Function updated successfully")
	}
	if result.Error == gorm.ErrRecordNotFound {
		errObj := models.ErrorResponse{Message: "Function does not exist"}
		return c.JSON(http.StatusBadRequest, errObj)
	} else {
		errObj := models.ErrorResponse{Message: result.Error.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}
}
