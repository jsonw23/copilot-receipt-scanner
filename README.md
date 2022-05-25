# AWS Copilot Sample Application - Receipt Scanner
This is the front-end API & app for my receipt scanner copilot sample application.  This needs to be installed first,
and then [copilot-receipt-scanner-backend](https://github.com/jsonw23/copilot-receipt-scanner-backend) is also needed.

## Installation
To install this on your AWS account, you'll need:
- AWS Credentials configured for programmatic access.  Install the AWS CLI and use the [configure command](https://docs.aws.amazon.com/cli/latest/reference/configure/)
- [AWS Copilot CLI](https://aws.amazon.com/containers/copilot/)
- Docker

The copilot manifest files are already in place, but you'll need to run `copilot init` to start provisioning resources and config in your AWS account.  Copilot will not read the manifest file before asking for service name and task type in the guided process, so pass some extra arguments for the service name and task type to match with the manifest.  Name it whatever you want.

```
copilot init -d ./Dockerfile -n api -t "Load Balanced Web Service" --deploy
```

After deployment is complete, you'll need to collect some values and bring them to the configuration for [copilot-receipt-scanner-backend](https://github.com/jsonw23/copilot-receipt-scanner-backend) because Copilot doesn't support cross-stack dependency injection.

- **The internet-facing load balancer url for the API**
- **The full name of the S3 bucket**
