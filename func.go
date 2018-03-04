package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func createEC2Session(profile string, region string) (*ec2.EC2, error) {
	p := Params{}
	if profile != "" {
		p.profile = profile
	}
	if region != "" {
		p.region = region
	}
	sess, err := newAwsSession(p)
	if err != nil {
		return nil, err
	}

	svc := ec2.New(sess)

	return svc, nil
}

func rmImages(svc *ec2.EC2, keep int, filters []string, dryrun bool, delss bool) error {
	images, err := findImages(svc, filters)
	if err != nil {
		return err
	}

	if len(images) == 0 {
		fmt.Printf("No images found.\n")
		return nil
	}
	if keep > len(images) {
		fmt.Printf("Number of keep is greater than number of images")
		return nil
	}

	// sort desc by CreationDate
	sort.Slice(images, func(i, j int) bool {
		return *images[i].CreationDate > *images[j].CreationDate
	})

	// split
	rmlist := images[keep:]

	for _, i := range rmlist {
		fmt.Printf("%s: %s, %s\n", *i.ImageId, *i.CreationDate, *i.Name)
		ssID := ""
		if delss {
			// find snapshot
			snapshots, err := findSnapshotsByImageID(svc, *i.ImageId)
			if err != nil {
				return err
			}
			if len(snapshots) > 1 {
				// TODO: err
				return fmt.Errorf("Multiple snapshots by %s\n", *i.ImageId)
			}
			tagName := findTagName(snapshots[0].Tags)
			fmt.Printf(" └─ %s: %s, %s, %s\n", *snapshots[0].SnapshotId, *snapshots[0].StartTime, *snapshots[0].VolumeId, tagName)
			ssID = *snapshots[0].SnapshotId
		}

		if dryrun {
			continue
		}
		err = removeImage(svc, *i.ImageId)
		if err != nil {
			return err
		}
		if delss {
			// remove snapshot
			err = removeSnapshot(svc, ssID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func findImages(svc *ec2.EC2, filters []string) ([]*ec2.Image, error) {
	params := &ec2.DescribeImagesInput{}
	awsFilters, err := buildFilters(filters)
	if err != nil {
		return nil, err
	}
	owners := make([]*string, 0)
	owners = append(owners, aws.String("self"))
	params.Owners = owners
	if len(awsFilters) > 0 {
		params.Filters = awsFilters
	}
	resp, err := svc.DescribeImages(params)
	if err != nil {
		return nil, err
	}

	return resp.Images, nil
}

func buildFilters(filters []string) ([]*ec2.Filter, error) {
	// filters=[]string{"Name1=Value11","Name2=Value21,Value22"}
	awsFilters := make([]*ec2.Filter, 0)
	for _, f := range filters {
		arr1 := strings.Split(f, "=")
		if len(arr1) != 2 {
			return nil, fmt.Errorf("filter:%s is invalid", f)
		}
		arr2 := strings.Split(arr1[1], ",")
		values := make([]*string, 0)
		for _, a := range arr2 {
			values = append(values, aws.String(a))
		}
		awsFilter := &ec2.Filter{
			Name:   aws.String(arr1[0]),
			Values: values,
		}

		awsFilters = append(awsFilters, awsFilter)
	}
	return awsFilters, nil
}

func removeImage(svc *ec2.EC2, imageID string) error {
	params := &ec2.DeregisterImageInput{
		ImageId: aws.String(imageID),
	}
	_, err := svc.DeregisterImage(params)
	if err != nil {
		return err
	}

	return nil
}

func findSnapshotsByImageID(svc *ec2.EC2, imageID string) ([]*ec2.Snapshot, error) {
	params := &ec2.DescribeSnapshotsInput{}
	values := make([]*string, 0)

	values = append(values, aws.String(fmt.Sprintf("Created by CreateImage(*) for %s *", imageID)))
	values = append(values, aws.String(fmt.Sprintf("Copied for DestinationAmi %s *", imageID)))

	awsFilters := make([]*ec2.Filter, 0)
	awsFilter := &ec2.Filter{
		Name:   aws.String("description"),
		Values: values,
	}
	awsFilters = append(awsFilters, awsFilter)
	params.Filters = awsFilters
	owners := make([]*string, 0)
	owners = append(owners, aws.String("self"))
	params.OwnerIds = owners
	resp, err := svc.DescribeSnapshots(params)
	// TODO: NextToken
	if err != nil {
		return nil, err
	}

	return resp.Snapshots, nil
}

func rmSnapshots(svc *ec2.EC2, keep int, filters []string, dryrun bool) error {
	snapshots, err := findSnapshots(svc, filters)
	if err != nil {
		return err
	}

	if len(snapshots) == 0 {
		fmt.Printf("No snapshots found.\n")
		return nil
	}
	if keep > len(snapshots) {
		fmt.Printf("Number of keep is greater than number of snapshots")
		return nil
	}

	// sort desc by StartTime
	sort.Slice(snapshots, func(i, j int) bool {
		return (*snapshots[i].StartTime).After(*snapshots[j].StartTime)
	})

	// split
	rmlist := snapshots[keep:]

	for _, i := range rmlist {
		tagName := findTagName(i.Tags)
		fmt.Printf("%s: %s, %s, %s\n", *i.SnapshotId, *i.StartTime, *i.VolumeId, tagName)
		if dryrun {
			continue
		}
		err = removeSnapshot(svc, *i.SnapshotId)
		if err != nil {
			return err
		}
	}

	return nil
}

func findSnapshots(svc *ec2.EC2, filters []string) ([]*ec2.Snapshot, error) {
	params := &ec2.DescribeSnapshotsInput{}
	awsFilters, err := buildFilters(filters)
	if err != nil {
		return nil, err
	}
	owners := make([]*string, 0)
	owners = append(owners, aws.String("self"))
	params.OwnerIds = owners
	if len(awsFilters) > 0 {
		params.Filters = awsFilters
	}
	resp, err := svc.DescribeSnapshots(params)
	// TODO: NextToken
	if err != nil {
		return nil, err
	}

	return resp.Snapshots, nil
}

func findTagName(tags []*ec2.Tag) string {
	for _, t := range tags {
		if *t.Key == "Name" {
			return *t.Value
		}
	}
	return ""
}

func removeSnapshot(svc *ec2.EC2, snapshotID string) error {
	params := &ec2.DeleteSnapshotInput{
		SnapshotId: aws.String(snapshotID),
	}
	_, err := svc.DeleteSnapshot(params)
	if err != nil {
		return err
	}

	return nil
}
