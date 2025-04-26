package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/lawn-chair/mealplan/utils"
)

func PostImageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	_, err := RequiresAuthentication(r)
	if err != nil {
		ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	storageEndpoint := utils.GetEnv("AWS_ENDPOINT_URL_S3", "localhost:9000")
	storageBucket := utils.GetEnv("BUCKET_NAME", "mp-images")

	minioClient, err := minio.New(storageEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(utils.GetEnv("AWS_ACCESS_KEY_ID", ""), utils.GetEnv("AWS_SECRET_ACCESS_KEY", ""), ""),
		Secure: false,
		Region: "auto",
	})
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	uuid := uuid.New()

	_, err = minioClient.PutObject(ctx,
		storageBucket,
		uuid.String()+fileHeader.Filename,
		file,
		fileHeader.Size,
		minio.PutObjectOptions{ContentType: "application/octet-stream"})

	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("Uploaded file: ", uuid.String()+fileHeader.Filename)
	json.NewEncoder(w).Encode(map[string]string{"url": "http://" + storageEndpoint + "/" + storageBucket + "/" + uuid.String() + fileHeader.Filename})
}
