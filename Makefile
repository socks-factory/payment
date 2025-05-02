NAME = socks-factory/payment
INSTANCE = payment

.PHONY: default copy test clean

default: payment

copy:
	docker create --name $(INSTANCE) $(NAME)-dev
	docker cp $(INSTANCE):/app $(shell pwd)/app
	docker rm $(INSTANCE)

release:
	docker build -t $(NAME) -f ./docker/payment/Dockerfile-release .

test:
	GROUP=weaveworksdemos COMMIT=$(COMMIT) ./scripts/build.sh
	./test/test.sh unit.py
	./test/test.sh container.py --tag $(COMMIT)

payment:
	CGO_ENABLED=0 go build -o payment cmd/paymentsvc/main.go

clean:
	rm payment
