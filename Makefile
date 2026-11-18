.PHONY: install-precommit fmt-check lint test build

install-precommit:
	@command -v lefthook >/dev/null 2>&1 || { \
		echo "lefthook is not installed. Install it with: go install github.com/evilmartians/lefthook@latest"; \
		exit 1; \
	}
	lefthook install

fmt-check:
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

lint:
	golangci-lint run ./...

test:
	go test -race ./...

build:
	go build ./cmd/...


# --- OpenTofu + c3x via Docker (no local installation required) ---
# Usage:
#   make tf-plan     TF_OPTION=option-b-lambda-functionurl TF_ENV=stage
#   make tf-estimate TF_ENV=prod

TF_IMAGE    ?= ghcr.io/opentofu/opentofu:1.9.0
C3X_IMAGE   ?= ghcr.io/c3xdev/c3x:latest
TF_PLATFORM ?= linux/amd64
TF_OPTION   ?= option-a-lambda-apigw
TF_ENV      ?= dev
AWS_REGION  ?= us-east-1

ROOT_DIR := $(shell git rev-parse --show-toplevel)

# The root module lives in terraform/<option>/. Per-environment settings
# (terraform.tfvars, backend.hcl) live under env/<env>/ and are referenced
# relative to that root module.
TF_ROOT      := terraform/$(TF_OPTION)
TF_DIR       := $(TF_ROOT)/env/$(TF_ENV)
TF_ENV_FILE  := env/$(TF_ENV)/terraform.tfvars
TF_BACKEND   := env/$(TF_ENV)/backend.hcl

# Common docker arguments. `-e VAR` forwards the host variable without a value,
# so credentials never appear in the Makefile or the process arguments.
TF_DOCKER_ARGS = --platform $(TF_PLATFORM) \
	-v "$(ROOT_DIR):/workspace" \
	-w "/workspace/$(TF_ROOT)" \
	-e AWS_ACCESS_KEY_ID \
	-e AWS_SECRET_ACCESS_KEY \
	-e AWS_SESSION_TOKEN \
	-e AWS_DEFAULT_REGION=$(AWS_REGION)

# Interactive runner (for apply/destroy confirmation prompts).
TF_RUN = docker run --rm -it $(TF_DOCKER_ARGS) $(TF_IMAGE)
# Non-interactive runner (clean stdout, safe in CI / when redirecting JSON).
TF_RUN_NOTTY = docker run --rm $(TF_DOCKER_ARGS) $(TF_IMAGE)

# c3x cost estimation, sharing the same workspace and working directory.
C3X_RUN = docker run --rm --platform $(TF_PLATFORM) \
	-v "$(ROOT_DIR):/workspace" \
	-w "/workspace/$(TF_ROOT)" \
	$(C3X_IMAGE)

.PHONY: tf-init tf-validate tf-plan tf-show tf-apply tf-destroy \
	tf-output tf-fmt tf-estimate tf-show-config

tf-init:
	$(TF_RUN) init -backend-config=$(TF_BACKEND)

tf-validate:
	$(TF_RUN) validate

# Write a saved plan to terraform/<option>/tfplan and display it.
tf-plan:
	$(TF_RUN) plan -var-file=$(TF_ENV_FILE) -out=tfplan
	$(TF_RUN_NOTTY) show tfplan

# Show the saved plan (run tf-plan or tf-estimate first).
tf-show:
	@if [ ! -f "$(ROOT_DIR)/$(TF_ROOT)/tfplan" ]; then \
		echo "No tfplan found. Run 'make tf-plan TF_OPTION=$(TF_OPTION) TF_ENV=$(TF_ENV)' first."; \
		exit 1; \
	fi
	$(TF_RUN_NOTTY) show tfplan

tf-apply:
	$(TF_RUN) apply -var-file=$(TF_ENV_FILE)

tf-destroy:
	$(TF_RUN) destroy -var-file=$(TF_ENV_FILE)

tf-output:
	$(TF_RUN) output

tf-fmt:
	docker run --rm --platform $(TF_PLATFORM) \
		-v "$(ROOT_DIR):/workspace" \
		-w "/workspace" \
		$(TF_IMAGE) fmt -recursive terraform

# Generate a plan + JSON and run c3x on that plan in the same command.
tf-estimate:
	$(TF_RUN_NOTTY) plan -var-file=$(TF_ENV_FILE) -out=tfplan
	$(TF_RUN_NOTTY) show -json tfplan > "$(ROOT_DIR)/$(TF_ROOT)/plan.json"
	$(C3X_RUN) estimate --path plan.json

tf-show-config:
	@echo "TF_OPTION   = $(TF_OPTION)"
	@echo "TF_ENV      = $(TF_ENV)"
	@echo "TF_ROOT     = $(TF_ROOT)"
	@echo "TF_DIR      = $(TF_DIR)"
	@echo "AWS_REGION  = $(AWS_REGION)"
	@echo "TF_IMAGE    = $(TF_IMAGE)"
	@echo "C3X_IMAGE   = $(C3X_IMAGE)"
	@echo "TF_PLATFORM = $(TF_PLATFORM)"
	@echo "ROOT_DIR    = $(ROOT_DIR)"
