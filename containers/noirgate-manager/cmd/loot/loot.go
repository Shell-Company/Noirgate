package loot

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	mySession = session.Must(session.NewSession())
)

func CreateTemporaryBucket(BucketName string) (BucketWebsite string) {
	publicWritePolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Sid":       "AddPerm",
				"Effect":    "Allow",
				"Principal": "*",
				"Action": []string{
					"s3:GetObject",
					"s3:PutObject",
					"s3:ListBucket",
					"s3:DeleteObject",
				},
				"Resource": []string{
					fmt.Sprintf("arn:aws:s3:::%s/*", BucketName),
					fmt.Sprintf("arn:aws:s3:::%s", BucketName),
				},
			},
		},
	}

	svc := s3.New(mySession, aws.NewConfig().WithRegion("us-west-2"))

	// Create S3 Bucket
	_, err := svc.CreateBucket(&s3.CreateBucketInput{Bucket: &BucketName})
	if err != nil {
		log.Println("Failed to create bucket", err)
	} else {
		log.Println("Created new bucket", BucketName)
	}

	// Put Bucket Website
	IndexDocument := "index.html"
	WebSiteConfig := &s3.WebsiteConfiguration{IndexDocument: &s3.IndexDocument{Suffix: &IndexDocument}}
	_, err = svc.PutBucketWebsite(&s3.PutBucketWebsiteInput{Bucket: &BucketName, WebsiteConfiguration: WebSiteConfig})
	if err != nil {
		log.Println("Failed to apply website configuration bucket ", err)
	} else {
		log.Println("Defined ", BucketName, " as an S3 website")
	}

	// Apply public write policy
	policy, _ := json.Marshal(publicWritePolicy)
	_, err = svc.PutBucketPolicy(&s3.PutBucketPolicyInput{
		Bucket: aws.String(BucketName),
		Policy: aws.String(string(policy)),
	})
	if err != nil {
		log.Println("Failed to create bucket ACL", err)
	} else {
		log.Println("Set acl to public-write on", BucketName)
	}

	// Apply bucket CORS policy
	sandboxCORSRule := s3.CORSRule{
		AllowedHeaders: aws.StringSlice([]string{"*"}),
		AllowedOrigins: aws.StringSlice([]string{"*"}),
		MaxAgeSeconds:  aws.Int64(600),

		// Add HTTP methods CORS request that were specified in the CLI.
		AllowedMethods: aws.StringSlice([]string{"POST", "GET", "HEAD", "DELETE"}),
		ExposeHeaders: aws.StringSlice([]string{
			"ETag",
			"x-amz-meta-custom-header",
			"x-amz-server-side-encryption",
			"x-amz-request-id",
			"x-amz-id-2",
			"date",
		}),
	}
	sandboxCORSparams := s3.PutBucketCorsInput{
		Bucket: &BucketName,
		CORSConfiguration: &s3.CORSConfiguration{
			CORSRules: []*s3.CORSRule{&sandboxCORSRule},
		},
	}

	_, err = svc.PutBucketCors(&sandboxCORSparams)
	if err != nil {
		// Print the error message
		log.Printf("Unable to set Bucket %q's CORS, %v", BucketName, err)
	}

	// Print the updated CORS config for the bucket
	log.Printf("Updated bucket %q CORS", BucketName)

	// http://noirgate-s3-sandbox-<id>.s3-website-us-west-2.amazonaws.com/ --no-sign-request
	return fmt.Sprintf("http://%s.s3-website-us-west-2.amazonaws.com ", BucketName)
}

func DeleteTemporaryBucket(BucketName string) {
	svc := s3.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
	// empty bucket contents

	// Setup BatchDeleteIterator to iterate through a list of objects.
	iter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
		Bucket: aws.String(BucketName),
	})

	// Traverse iterator deleting each object
	if err := s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), iter); err != nil {
		log.Printf("Unable to delete objects from bucket %q, %v", BucketName, err)
	}

	log.Printf("Deleted object(s) from bucket: %s", BucketName)
	// Delete bucket
	_, err := svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: &BucketName,
	})
	if err != nil {
		log.Fatal("Failed to delete bucket", err)
	} else {
		log.Println("Deleted bucket", BucketName)
	}
}
