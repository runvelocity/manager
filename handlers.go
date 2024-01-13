package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/labstack/echo"
	"gorm.io/gorm"
)

func GetFunctionsHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	var functions []Function
	result := db.Find(&functions)
	if result.Error != nil {
		errObj := ErrorResponse{Message: result.Error.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	functionsResponse := FunctionsResponse{Functions: functions}
	return c.JSON(http.StatusOK, functionsResponse)
}

// getFunction is the HTTP handler for GET /functions/{uuid}.
func GetFunctionHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	uuid := c.Param("uuid")
	var function Function

	result := db.Where("uuid = ?", uuid).First(&function)
	if result.Error != nil {
		errObj := ErrorResponse{Message: result.Error.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	return c.JSON(http.StatusOK, function)
}

// createFunction is the HTTP handler for POST /functions.
func CreateFunctionHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	var function Function
	err := c.Bind(&function)
	if err != nil {
		errObj := ErrorResponse{Message: "Error while reading request body"}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	function.Status = PENDING
	var resObj Function
	result := db.Where("name = ?", function.Name).First(&resObj)

	if result.Error == nil {
		errObj := ErrorResponse{Message: "A function with this name already exists"}
		return c.JSON(http.StatusBadRequest, errObj)
	}
	if result.Error == gorm.ErrRecordNotFound {
		db.Create(function)
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		sfnArn := os.Getenv("STEP_FUNCTIONS_ARN")

		sfnRequest := StepFunctionsRequest{
			UUID:         function.UUID,
			CodeLocation: function.CodeLocation,
		}

		svc := sfn.New(sess)
		bytes, err := json.Marshal(sfnRequest)
		if err != nil {
			errObj := ErrorResponse{Message: err.Error()}
			return c.JSON(http.StatusInternalServerError, errObj)
		}
		input := string(bytes)
		_, err = svc.StartExecution(&sfn.StartExecutionInput{
			Input:           aws.String(input),
			StateMachineArn: aws.String(sfnArn),
		})

		if err != nil {
			errObj := ErrorResponse{Message: err.Error()}
			return c.JSON(http.StatusInternalServerError, errObj)
		}
		return c.JSON(http.StatusCreated, function)
	} else {
		errObj := ErrorResponse{Message: result.Error.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}
}

// getFunction is the HTTP handler for DELETE /functions/{name}.
func DeleteFunctionHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	uuid := c.Param("uuid")
	var function Function
	result := db.Where("uuid = ?", uuid).First(&function)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.String(http.StatusNotFound, "Function does not exist")
		} else {
			errObj := ErrorResponse{Message: result.Error.Error()}
			return c.JSON(http.StatusInternalServerError, errObj)
		}
	}
	result = db.Where("uuid = ?", uuid).Delete(&function)
	if result.Error != nil {
		errObj := ErrorResponse{Message: result.Error.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	return c.String(http.StatusOK, "Function deleted successfully")
}

func UpdateFunctionHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	var function Function
	err := c.Bind(&function)
	if err != nil {
		errObj := ErrorResponse{Message: "Error while reading request body"}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	uuid := c.Param("uuid")
	var resObj Function
	result := db.Where("uuid = ?", uuid).First(&resObj)

	if result.Error == nil {
		db.Where("uuid = ?", uuid).Model(&resObj).Updates(function)
		return c.String(http.StatusOK, "Function updated successfully")
	}
	if result.Error == gorm.ErrRecordNotFound {
		errObj := ErrorResponse{Message: "Function does not exist"}
		return c.JSON(http.StatusBadRequest, errObj)
	} else {
		errObj := ErrorResponse{Message: result.Error.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}
}

func UploadHandler(c echo.Context) error {
	var uploadArgs UploadHandlerArgs
	err := c.Bind(&uploadArgs)
	if err != nil {
		errObj := ErrorResponse{Message: "Error while reading request body"}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	formFile, err := c.FormFile("code")
	if err != nil {
		fmt.Println(err.Error())
		errObj := ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	file, err := formFile.Open()
	if err != nil {
		fmt.Println(err.Error())
		errObj := ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}
	if err != nil {
		errObj := ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	uploader := s3manager.NewUploader(sess)

	obj, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("S3_UPLOAD_BUCKET")),
		Key:    aws.String(fmt.Sprintf("code/%s.zip", uploadArgs.Key)),
		Body:   file,
	})

	if err != nil {
		errObj := ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}
	return c.JSON(http.StatusOK, obj)
}

// getFunction is the HTTP handler for GET /invoke.
func InvokeHandler(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	worker_url := os.Getenv("WORKER_URL")
	name := c.Param("name")
	var function Function
	var args map[string]interface{}
	err := c.Bind(&args)
	if err != nil {
		errObj := ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}
	result := db.Where("name = ?", name).First(&function)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.String(http.StatusNotFound, "Function does not exist")
		} else {
			errObj := ErrorResponse{Message: result.Error.Error()}
			return c.JSON(http.StatusInternalServerError, errObj)
		}
	}
	invokePayload := InvokePayload{
		Handler: function.Handler,
		Args:    args,
	}

	invokeRequest := InvokeRequest{
		VmId:          function.UUID,
		InvokePayload: invokePayload,
	}

	argsJSON, err := json.Marshal(invokeRequest)
	if err != nil {
		fmt.Println("Error marshaling args to JSON:", err)
		errObj := ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	req, err := http.NewRequest("POST", worker_url+"/invoke", bytes.NewBuffer(argsJSON))
	if err != nil {
		fmt.Println("Error creating request:", err)
		errObj := ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	// Set headers if needed
	req.Header.Set("Content-Type", "application/json")

	req.Close = true

	var client = &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		errObj := ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}
	var resp map[string]interface{}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err.Error())
		errObj := ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		fmt.Println(err.Error())
		errObj := ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	return c.JSON(res.StatusCode, resp)
}
