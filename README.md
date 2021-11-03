# Microservices

A basic example of microservice architecture which demonstrates communication between a few loosely coupled services.

* Written in Go
* Uses RabbitMQ to communicate between services
* Uses WebSocket to talk to the front end
* Stores data in PostgreSQL
* Stores cache in Redis
* Uses React for front end development
* Builds and runs with Docker
* Deployed on AWS with CloudFormation templates
    * ECS using EC2
    * AWS Fargate

![](demo.gif)

## Deployment on AWS Fargate

`cd deployments` and create the pipeline stack:

```bash
aws cloudformation create-stack \
	--stack-name MicroservicesPipeline \
	--template-body file://pipeline.yml \
	--parameters \
		ParameterKey=DeploymentType,ParameterValue=fargate \
		ParameterKey=EnvironmentName,ParameterValue=microservices \
		ParameterKey=GitHubRepo,ParameterValue=<github_repo_name> \
		ParameterKey=GitHubBranch,ParameterValue=<github_branch> \
		ParameterKey=GitHubToken,ParameterValue=<github_token> \
		ParameterKey=GitHubUser,ParameterValue=<github_user> \
	--capabilities CAPABILITY_NAMED_IAM
```

## Local use

To run the example, clone the Github repository and start the services using Docker Compose. Once Docker finishes downloading and building images, the front end is accessible by visiting `localhost:8080`.

```bash
git clone https://github.com/ebosas/microservices
cd microservices
```
```bash
docker-compose up
```

### Database

To inspect the database, launch a new container that will connect to our Postgres database. Then enter the password `demopsw` (see the `.env` file).

```bash
docker run -it --rm \
    --network microservices_network \
    postgres:13-alpine \
    psql -h postgres -U postgres -d microservices
```

Select everything from the messages table:

```sql
select * from messages;
```

### Redis

To inspect Redis, connect to its container via redis-cli.

```bash
docker run -it --rm \
    --network microservices_network \
    redis:6-alpine \
    redis-cli -h redis
```

Get all cached messages or show the number of messages.

```bash
lrange messages 0 -1
get total
```

### RabbitMQ

Access the RabbitMQ management interface by visiting `localhost:15672` with `guest` as both username and password.

### Back end

To access the back end service, attach to its docker container from a separate terminal window. Messages from the front end will show up here. Also, standart input will be sent to the front end for two way communication.

```bash
docker attach microservices_backend
```

## Deployment to Amazon ECS/AWS Fargate

### ECR

To deploy on ECS, the first step is to publish the service images to Amazon ECR. In Makefile, replace the `registry` variable with your own registry. [Login](https://docs.aws.amazon.com/AmazonECR/latest/userguide/registry_auth.html#get-login-password) to ECR. Then run the following command:

```bash
make ecr
```

### CloudFormation stacks

There are several CloudFormation stacks to create. `cd deployments` and create the network stack:

```bash
aws cloudformation create-stack --stack-name MicroservicesNetwork --template-body file://network.yml
```

Create the following stacks in any order (for an EC2 cluster). To create a Fargate cluster, change the `cluster-ec2.yml` to `cluster-fargate.yml`.

```bash
aws cloudformation create-stack --stack-name MicroservicesResources --template-body file://resources.yml
aws cloudformation create-stack --stack-name MicroservicesAlb --template-body file://alb.yml
aws cloudformation create-stack --stack-name MicroservicesClusterEC2 --template-body file://cluster-ec2.yml --capabilities CAPABILITY_NAMED_IAM
```

Once done, create the services (for an EC2 cluster). For Fargate, change the `services-ec2` directory to `services-fargate`.

```bash
aws cloudformation create-stack --stack-name MicroservicesServiceServer --template-body file://services-ec2/server.yml
aws cloudformation create-stack --stack-name MicroservicesServiceCache --template-body file://services-ec2/cache.yml
aws cloudformation create-stack --stack-name MicroservicesServiceDatabase --template-body file://services-ec2/database.yml
```

To visit the website, head to the MicroservicesAlb stack and open the `ExternalUrl` from the Outputs tab.

### Deleting stacks

When deleting the `cluster-ec2.yml` stack in CloudFormation, delete the auto scaling group manually from the AWS EC2 console.

### References

Deployment is based on these templates: https://github.com/nathanpeck/ecs-cloudformation

## Development environment

For development, run the RabbitMQ and Postgres containers with Docker Compose.

```bash
docker-compose -f docker-compose-dev.yml up
```

Generate static web assets for the server service by going to `web/react` and `web/bootstrap` and running:

```bash
npm run build-server
```

### React

For React development, run `npm run serve` in `web/react` and change the script tag in the server's template to the following:

```html
<script src="http://127.0.0.1:8000/index.js"></script>
```
