package service

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/url"
	"os"
	"strings"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

const (
	cadFileContainer = "fxtcadfiles"
	pdfContainer     = "fxtpdfs"

	PDFFILE = 0
	CADFILE = 1
)

type FILETYPE int

type AzureBlobService interface {
	UploadFromBuffer(b *bytes.Buffer, filename string) (azblob.CommonResponse, string, error)
	UploadFromFile(file *multipart.File, filename string) (azblob.CommonResponse, string, error)
	Delete(url string, ftype FILETYPE) (*azblob.BlobDeleteResponse, error)
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

func (a *azureBlobService) UploadFromFile(file *multipart.File, filename string) (azblob.CommonResponse, string, error) {
	cURL := a.serviceURL.NewContainerURL(cadFileContainer)
	bURL := cURL.NewBlockBlobURL(filename)

	resp, err := azblob.UploadStreamToBlockBlob(context.Background(), *file, bURL, azblob.UploadStreamToBlockBlobOptions{})
	if err != nil {
		return nil, "", err
	}

	return resp, bURL.String(), nil
}

func (a *azureBlobService) Delete(fileURL string, ftype FILETYPE) (*azblob.BlobDeleteResponse, error) {
	var cURL azblob.ContainerURL

	if ftype == CADFILE {
		cURL = a.serviceURL.NewContainerURL(cadFileContainer)
	} else {
		cURL = a.serviceURL.NewContainerURL(pdfContainer)
	}

	link, _ := url.Parse(fileURL)
	urlParts := strings.Split(link.Path, "/")

	bURL := cURL.NewBlockBlobURL(fmt.Sprintf("%s/%s", urlParts[2], urlParts[3]))

	resp, err := bURL.Delete(context.Background(), azblob.DeleteSnapshotsOptionInclude, azblob.BlobAccessConditions{})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
