FUNCTIONS_DIR = ./functions
build: $(FUNCTIONS_DIR)/*
			for file in $^; do \
				GOARCH=amd64 GOOS=linux go build -o ./dist/$${file}/main $${file}/main.go; \
			done
