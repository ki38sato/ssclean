ssclean
===
AWS AMI/Snapshot cleaner


CommandLine Option
---
|option key|description|
|---|---|
|del-snapshot|(kind=ami only) delete source snapshot with ami togetther|
|dryrun|dryrun|
|filters|use filters. see [doc 1](https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#DescribeImagesInput) [doc 2](https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#DescribeSnapshotsInput)|
|keep|number of age remaining|
|kind|(required) ami or snapshot|
|profile|specify aws credential profile|
|region|specify aws region|


Example Usage
---
AMI
```
$ ssclean --region ap-northeast-1 --profile yourProfile --kind ami --filters "name=AMI-*" --keep 5 --del-snapshot --dryrun
```

Snapshot
```
$ ssclean --kind snapshot --keep 7 --dryrun
```

FEATURE
---
- more example (filters..)
- describe-snapshots nextToken loop
- older-than option


LICENSE
---
MIT
