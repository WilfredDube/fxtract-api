package service

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

const (
	cadFileContainer = "fxtcadfiles"
	pdfContainer     = "fxtpdfs"
)

type AzureBlobService interface {
	UploadFromBuffer(b *bytes.Buffer, filename string) (azblob.CommonResponse, string, error)
	UploadFromFile(file *multipart.File, path, folder string, filename int64) (azblob.CommonResponse, string, error)
}

type azureBlobService struct {
	serviceURL azblob.ServiceURL
}

func NewAzureBlobService() AzureBlobService {
	accountName := os.Getenv("AZURE_BLOB_STORAGE_NAME")
	accountKey := os.Getenv("AZURE_BLOB_STORAGE_KEY")
	storageURL := os.Getenv("AZURE_BLOB_STORAGE_URL")

	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatalln("Not able to connect to storage account")
	}

	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	p := azblob.NewPipeline(cred, azblob.PipelineOptions{})

	u, err := url.Parse(fmt.Sprintf(storageURL, accountName))
	if err != nil {
		log.Fatalln("Not able to connect to storage account")
	}

	serviceURL := azblob.NewServiceURL(*u, p)
	return &azureBlobService{serviceURL: serviceURL}
}

func (a *azureBlobService) UploadFromBuffer(buf *bytes.Buffer, filename string) (azblob.CommonResponse, string, error) {
	cURL := a.serviceURL.NewContainerURL(pdfContainer)
	bURL := cURL.NewBlockBlobURL(filename)

	resp, err := azblob.UploadBufferToBlockBlob(context.Background(), buf.Bytes(), bURL, azblob.UploadToBlockBlobOptions{})
	if err != nil {
		return nil, "", err
	}

	return resp, bURL.String(), nil
}

func (a *azureBlobService) UploadFromFile(file *multipart.File, path, folder string, filename int64) (azblob.CommonResponse, string, error) {
	cURL := a.serviceURL.NewContainerURL(cadFileContainer)
	bURL := cURL.NewBlockBlobURL(fmt.Sprintf(folder+"/%d%s", filename, filepath.Ext(path)))

	resp, err := azblob.UploadStreamToBlockBlob(context.Background(), *file, bURL, azblob.UploadStreamToBlockBlobOptions{})
	if err != nil {
		return nil, "", err
	}

	return resp, bURL.String(), nil
}
