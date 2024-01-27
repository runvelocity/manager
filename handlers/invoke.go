package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/runvelocity/manager/models"
	"gorm.io/gorm"
)

func InvokeHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	worker_url := os.Getenv("WORKER_URL")
	name := c.Param("name")
	var function models.Function
	var args map[string]interface{}
	err := c.Bind(&args)
	if err != nil {
		errObj := models.ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}
	result := db.Where("name = ?", name).First(&function)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.String(http.StatusNotFound, "Function does not exist")
		} else {
			errObj := models.ErrorResponse{Message: result.Error.Error()}
			return c.JSON(http.StatusInternalServerError, errObj)
		}
	}
	invokePayload := models.InvokePayload{
		Handler: function.Handler,
		Args:    args,
	}

	invokeRequest := models.InvokeRequest{
		FunctionId:    function.UUID,
		InvokePayload: invokePayload,
	}

	argsJSON, err := json.Marshal(invokeRequest)
	if err != nil {
		fmt.Println("Error marshaling args to JSON:", err)
		errObj := models.ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	req, err := http.NewRequest("POST", worker_url+"/invoke", bytes.NewBuffer(argsJSON))
	if err != nil {
		fmt.Println("Error creating request:", err)
		errObj := models.ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	req.Header.Set("Content-Type", "application/json")

	req.Close = true

	var client = &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		errObj := models.ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}
	var resp map[string]interface{}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err.Error())
		errObj := models.ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		fmt.Println(err.Error())
		errObj := models.ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	return c.JSON(res.StatusCode, resp)
}
