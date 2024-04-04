EXAMPLES_DIR:=./examples

EXAMPLES_GENERATE_COMMON_QUERIES_CONF:=$(EXAMPLES_DIR)/common/sqlc.yaml
EXAMPLES_COMMON_MIGRATIONS:=$(shell find $(EXAMPLES_DIR)/common/migrations -type f)
EXAMPLES_COMMON_QUERIES:=$(shell find $(EXAMPLES_DIR)/common/queries -type f)
EXAMPLES_COMMON_GENERATED_QUERIES_DIR:=$(EXAMPLES_DIR)/common/dbqueries


.PHONY: default generate lint common_example_queries test_lib test_examples test ci

default: generate

lint:
	golangci-lint run ./... $(EXAMPLES_DIR)/...


$(EXAMPLES_COMMON_GENERATED_QUERIES_DIR): $(EXAMPLES_GENERATE_COMMON_QUERIES_CONF) $(EXAMPLES_COMMON_MIGRATIONS) $(EXAMPLES_COMMON_QUERIES)
	sqlc -f $(EXAMPLES_GENERATE_COMMON_QUERIES_CONF) generate

common_example_queries: $(EXAMPLES_COMMON_GENERATED_QUERIES_DIR)

generate: common_example_queries

test_lib:
	go test ./... -cover -race -count 1

test_examples:
	docker run -it --rm --name 'pg_15_test' -e POSTGRES_USER=localuser -e POSTGRES_PASSWORD=localpa55w.rd -d -p 5400:5432 postgres:15
	sleep 5
	PGTEST_HOST='localhost' \
		    PGTEST_PORT=5400 \
		    PGTEST_USER='localuser' \
		    PGTEST_PASSWORD='localpa55w.rd' \
		    go test $(EXAMPLES_DIR)/... -cover -race -count 1 || EXIT_CODE=$$?; \
		    docker container stop pg_15_test; \
		    exit $$EXIT_CODE

test: test_lib test_examples

ci: lint test
