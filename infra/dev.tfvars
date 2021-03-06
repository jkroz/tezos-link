REGION = "eu-west-1"

ENV = "dev"

VPC_CIDR                        = "10.1.0.0/16"

SUBNET_TZ_FARM_CIDR             = "10.1.0.0/24"
SUBNET_TZ_PUBLIC_PROXY_CIDR     = "10.1.1.0/24"
SUBNET_TZ_PRIVATE_PROXY_CIDR    = "10.1.2.0/24"
SUBNET_TZ_PRIVATE_DATABASE_CIDR = "10.1.3.0/24"

DATABASE_USERNAME = "administrator"

NODE_USER = "ec2-user"

SNAPSHOT_S3_KEY = "v1.0.0/snapshot.zip"

S3_LAMBDA_KEY = "snapshot_lambda_key"
