package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/labstack/echo"
	"github.com/runvelocity/manager/models"
)

func UploadHandler(c echo.Context) error {
	var uploadArgs models.UploadHandlerArgs
	err := c.Bind(&uploadArgs)
	if err != nil {
		errObj := models.ErrorResponse{Message: "Error while reading request body"}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	formFile, err := c.FormFile("code")
	if err != nil {
		fmt.Println(err.Error())
		errObj := models.ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}

	file, err := formFile.Open()
	if err != nil {
		fmt.Println(err.Error())
		errObj := models.ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}
	if err != nil {
		errObj := models.ErrorResponse{Message: err.Error()}
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
		errObj := models.ErrorResponse{Message: err.Error()}
		return c.JSON(http.StatusInternalServerError, errObj)
	}
	return c.JSON(http.StatusOK, obj)
}
