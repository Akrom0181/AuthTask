package helper

// import (
// 	"context"
// 	"fmt"
// 	"barber/api/models"
// 	"io"
// 	"log"
// 	"mime/multipart"
// 	"net/url"
// 	"path/filepath"

// 	firebase "firebase.google.com/go"
// 	"github.com/google/uuid"
// 	"google.golang.org/api/option"
// 	"google.golang.org/api/storage/v1"
// )

// // UploadFiles uploads multiple files to Firebase Storage and returns their URLs.
// func UploadFiles(file *multipart.Form) (*models.MultipleFileUploadResponse, error) {
// 	var resp models.MultipleFileUploadResponse

// 	filePath := filepath.Join("./", "serviceAccountKey.json")
// 	opt := option.WithCredentialsFile(filePath)
// 	app, err := firebase.NewApp(context.Background(), nil, opt)
// 	if err != nil {
// 		log.Println("Firebase App initialization error:", err)
// 		return nil, err
// 	}
// 	client, err := app.Storage(context.Background())
// 	if err != nil {
// 		log.Println("Firebase Storage client initialization error:", err)
// 		return nil, err
// 	}

// 	bucketHandle, err := client.Bucket("food-8ceb4.appspot.com")
// 	if err != nil {
// 		log.Println("Bucket handle error:", err)
// 		return nil, err
// 	}

// 	for _, v := range file.File["file"] {
// 		id := uuid.New().String()
// 		imageFile, err := v.Open()
// 		if err != nil {
// 			log.Println("Error opening file:", v.Filename, err)
// 			return nil, err
// 		}
// 		defer imageFile.Close() 

// 		log.Println("Uploading file:", v.Filename)

// 		objectHandle := bucketHandle.Object(v.Filename)
// 		writer := objectHandle.NewWriter(context.Background())
// 		writer.ObjectAttrs.Metadata = map[string]string{"firebaseStorageDownloadTokens": id}

// 		if _, err := io.Copy(writer, imageFile); err != nil {
// 			log.Printf("Error copying file %s to Firebase Storage: %v", v.Filename, err)
// 			return nil, err
// 		}

// 		if err := writer.Close(); err != nil {
// 			log.Printf("Error closing writer for file %s: %v", v.Filename, err)
// 			return nil, err
// 		}

// 		log.Println("File uploaded successfully:", v.Filename)

// 		encodedFileName := url.PathEscape(v.Filename)
// 		fileURL := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/food-8ceb4.appspot.com/o/%s?alt=media&token=%s", encodedFileName, id)

// 		resp.Url = append(resp.Url, &models.Url{
// 			Id:  id,
// 			Url: fileURL,
// 		})
// 	}

// 	return &resp, nil
// }

// func DeleteFile(id string) error {
// 	ctx := context.Background()
// 	client, err := storage.NewService(ctx, option.WithCredentialsFile("serviceAccountKey.json"))
// 	if err != nil {
// 		log.Println("Failed to create client:", err)
// 		return err
// 	}

// 	// Bucket name and object path to delete
// 	bucketName := "food-8ceb4.appspot.com"
// 	objectPath := id

// 	// Delete the object
// 	err = client.Objects.Delete(bucketName, objectPath).Do()
// 	if err != nil {
// 		log.Println("Failed to delete object:", err)
// 		return err
// 	}

// 	fmt.Printf("Object %s deleted successfully from bucket %s\n", objectPath, bucketName)
// 	return nil
// }
