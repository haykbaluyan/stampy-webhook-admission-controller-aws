package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
)

// ImageControllerInterface exposes image related operations
type ImageControllerInterface interface {
	GetManifest(string, string) (string, error)
	GetManifestSignature(string, string) (string, error)
}

// imageController implements image related operations for AWS
type imageController struct {
	region string
	bucket string
	logger *logrus.Logger
}

// NewImageController constructor
func NewImageController(region, bucket string, logger *logrus.Logger) ImageControllerInterface {
	return &imageController{
		region: region,
		bucket: bucket,
		logger: logger,
	}
}

// GetManifestSignature returns manifest signature of the image
func (aim *imageController) GetManifestSignature(repo, digest string) (string, error) {
	manifestSigURL := fmt.Sprintf("%s/%s/manifest.json.sig", repo, digest)
	sess, err := aim.createSession()
	if err != nil {
		return "", errors.Trace(err)
	}

	buf := aws.NewWriteAtBuffer([]byte{})
	downloader := s3manager.NewDownloader(sess)
	_, err = downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(aim.bucket),
			Key:    aws.String(manifestSigURL),
		})
	if err != nil {
		errors.Errorf("api=GetManifestSignature, manifestSigURL=%q, err=%v", manifestSigURL, err)
	}

	return string(buf.Bytes()), nil
}

// GetManifest returns manifest of the image
func (aim *imageController) GetManifest(repo, tag string) (string, error) {
	sess, err := aim.createSession()
	if err != nil {
		return "", errors.Trace(err)
	}

	ecrSvc := ecr.New(sess)
	inputBatchGetImage := &ecr.BatchGetImageInput{
		ImageIds: []*ecr.ImageIdentifier{
			{
				ImageTag: aws.String(tag),
			},
		},
		RepositoryName: aws.String(repo),
	}

	resultBatchGetImage, err := ecrSvc.BatchGetImage(inputBatchGetImage)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ecr.ErrCodeServerException:
				return "", errors.Errorf("api=BatchGetImage, awsErrCode=%q, err=%v", ecr.ErrCodeServerException, aerr.Error())
			case ecr.ErrCodeInvalidParameterException:
				return "", errors.Errorf("api=BatchGetImage, awsErrCode=%q, err=%v", ecr.ErrCodeInvalidParameterException, aerr.Error())
			case ecr.ErrCodeRepositoryNotFoundException:
				return "", errors.Errorf("api=BatchGetImage, awsErrCode=%q, err=%v", ecr.ErrCodeRepositoryNotFoundException, aerr.Error())
			default:
				return "", errors.Errorf("api=BatchGetImage, err=%v", aerr.Error())
			}
		}
		return "", errors.Errorf("api=BatchGetImage, repo=%q, tag=%q, err=%v", repo, tag, err)
	}

	if len(resultBatchGetImage.Images) == 0 {
		return "", errors.Errorf("api=GetManifest, reason='failed to get any image with the specified parameters' repo=%q, tag=%q", repo, tag)
	}

	if len(resultBatchGetImage.Images) > 1 {
		return "", errors.Errorf("api=GetManifest, reason='more than one image found with the specified parameters' repo=%q, tag=%q", repo, tag)
	}

	return *resultBatchGetImage.Images[0].ImageManifest, nil
}

func (aim *imageController) createSession() (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(aim.region),
	})

	return sess, err
}
