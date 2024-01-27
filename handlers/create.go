package handlers

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/runvelocity/manager/models"
	"gorm.io/gorm"
)

// createFunction is the HTTP handler for POST /functions.
func CreateFunctionHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	var function models.Function
	err := c.Bind(&function)
	if err != nil {
		errObj := models.ErrorResponse{Message: "Error while reading request body"}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	var resObj models.Function
	result := db.Where("name = ?", function.Name).First(&resObj)

	if result.Error == nil {
		errObj := models.ErrorResponse{Message: "A function with this name already exists"}
		return c.JSON(http.StatusBadRequest, errObj)
	}
	if result.Error == gorm.ErrRecordNotFound {
		db.Create(function)
		return c.JSON(http.StatusCreated, function)
	} else {
		errObj := models.ErrorResponse{Message: result.Error.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}
}
