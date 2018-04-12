package main

import (
	"context"
	"log"
	"os"
	"fmt"
	"strconv"
	"encoding/json"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
)

type Ec2EventDetail struct {
	InstanceId string `json:"instance-id"`
	State string `json:"state"`
}

func GetDescribeInstances(d Ec2EventDetail) () {
	r := os.Getenv("REGION")
	if r == "" {
		log.Print("Undefined Env REGION.")
		return
	}

	tk := os.Getenv("TAGKEY")
	if tk == "" {
		log.Print("Undefined Env TAGKEY.")
		return
	}

	f := ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			&d.InstanceId,
		},
	}
	svc := ec2.New(session.New(), aws.NewConfig().WithRegion(r))
	res, err := svc.DescribeInstances(&f)
	if err != nil {
		log.Print("Get Describe Error.")
		return
	}

	for _, r := range res.Reservations {
		for _, i := range r.Instances {
			for _, t := range i.Tags {
				if *t.Key == tk {
					AddRoute53(t.Value, i.PrivateIpAddress)
				}
			}
		}
	}

}

func AddRoute53(v *string, ip *string) () {
	svc := route53.New(session.New(), aws.NewConfig())

	act := "CREATE"
	rtype := "A"

	ttlenv, err := strconv.Atoi(os.Getenv("TTL"))
	if err != nil {
		log.Print("Undefined Env TTL.")
		return
	}
	ttl := int64(ttlenv)

	hzone := os.Getenv("HOSTZONE")
	if hzone == "" {
		log.Print("Undefined Env HOSTZONE.")
		return
	}

	dname := os.Getenv("DOMAIN")
	if dname == "" {
		log.Print("Undefined Env DOMAIN.")
		return
	}
	hname := *v + "." + dname

	recordset := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(act),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(hname),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: ip,
							},
						},
						TTL:  aws.Int64(ttl),
						Type: aws.String(rtype),
					},
				},
			},
		},
		HostedZoneId: aws.String(hzone),
	}

	res, err := svc.ChangeResourceRecordSets(recordset)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case route53.ErrCodeNoSuchHostedZone:
				fmt.Println(route53.ErrCodeNoSuchHostedZone, aerr.Error())
			case route53.ErrCodeNoSuchHealthCheck:
				fmt.Println(route53.ErrCodeNoSuchHealthCheck, aerr.Error())
			case route53.ErrCodeInvalidChangeBatch:
				fmt.Println(route53.ErrCodeInvalidChangeBatch, aerr.Error())
			case route53.ErrCodeInvalidInput:
				fmt.Println(route53.ErrCodeInvalidInput, aerr.Error())
			case route53.ErrCodePriorRequestNotComplete:
				fmt.Println(route53.ErrCodePriorRequestNotComplete, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return
	}
	log.Print("" + res.ChangeInfo.Id)
	return
}

func Handler(ctx context.Context, event events.CloudWatchEvent) () {
	var detail Ec2EventDetail
	err := json.Unmarshal(event.Detail, &detail)
	if err != nil {
		fmt.Printf("%#v\n", err)
		return
	}
	log.Print("Start.")

	GetDescribeInstances(detail)

	log.Print("Finish.")
	return

}

func main() {
	lambda.Start(Handler)
}