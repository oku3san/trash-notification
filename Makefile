include .envrc

.PHONY: init-local
init-local:
	docker compose -f ./src/docker-compose.yml down
	docker compose -f ./src/docker-compose.yml up -d
	sleep 5
	make setup-domain-local
	pushd infra && \
		cdklocal context --clear -c env=local && \
		cdklocal bootstrap -c env=local && \
		cdklocal deploy -c env=local --require-approval never && \
		popd
	make setup-data-local

.PHONY: setup-data-local
setup-data-local:
		$(eval TABLENAME := $(shell awslocal cloudformation describe-stacks --region ap-northeast-1 --stack-name "TrashNotificationStack" --output text --query 'Stacks[].Outputs[?OutputKey==`dynamoDbName`].[OutputValue]'))
		gsed "s/<TableName>/$(TABLENAME)/g" ./src/dynamodb/seed-template.json > ./src/dynamodb/seed.json && \
		awslocal dynamodb batch-write-item --request-items file://./src/dynamodb/seed.json --region ap-northeast-1 ; \
		rm ./src/dynamodb/seed.json

.PHONY: setup-domain-local
setup-domain-local:
		awslocal route53 create-hosted-zone --name ${domainName} --caller-reference `date +%Y-%m-%d_%H-%M-%S` --region ap-northeast-1

.PHONY: setup-data
setup-data:
		$(eval TABLENAME := $(shell aws-vault exec default -- aws cloudformation describe-stacks --region ap-northeast-1 --stack-name "TrashNotificationStack" --output text --query 'Stacks[].Outputs[?OutputKey==`dynamoDbName`].[OutputValue]'))
		gsed "s/<TableName>/$(TABLENAME)/g" ./src/dynamodb/seed-template.json > ./src/dynamodb/seed.json && \
		aws-vault exec default -- aws dynamodb batch-write-item --request-items file://./src/dynamodb/seed.json --region ap-northeast-1 ; \
		rm ./src/dynamodb/seed.json

.PHONY: deploy-local
deploy-local:
	pushd infra && \
		cdklocal deploy -c env=local --require-approval never && \
		popd

.PHONY: deploy
deploy:
	pushd infra && \
		aws-vault exec default -- cdk deploy -c env=prd --require-approval never && \
		popd

.PHONY: diff-local
diff-local:
	pushd infra && \
		cdklocal diff -c env=local && \
		popd

.PHONY: diff
diff:
	pushd infra && \
		aws-vault exec default -- cdk diff -c env=prd && \
		popd

.PHONY: send-message-local
send-message-local:
		$(eval SQSENDPOINT := $(shell awslocal cloudformation describe-stacks --region ap-northeast-1 --stack-name "TrashNotificationStack" --output text --query 'Stacks[].Outputs[?OutputKey==`sqsEndpoint`].[OutputValue]'))
		awslocal sqs send-message --message-body '{"destination":"xxx","events":[{"type":"message","message":{"type":"text","id":"xxx","text":"$(MSG)"},"webhookEventId":"xxx","deliveryContext":{"isRedelivery":false},"timestamp":1672845789652,"source":{"type":"user","userId":"xxx"},"replyToken":"xxx","mode":"active"}]}' --region ap-northeast-1 --queue-url $(SQSENDPOINT)
