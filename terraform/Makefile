
# --- OpenTofu + C3X via Docker ---
# Run from terraform/ or use make -f terraform/Makefile from repo root.
#
# Examples:
#   make tf-init TF_OPTION=option-a-lambda-apigw TF_ENV=prod
#   make tf-plan TF_OPTION=option-b-lambda-functionurl TF_ENV=stage
#   make tf-estimate TF_OPTION=option-a-lambda-apigw TF_ENV=prod
#   make tf-show
#
# AWS credentials are read from the host environment or ~/.aws.

# Images
TF_IMAGE      ?= ghcr.io/opentofu/opentofu:1.9.0
C3X_IMAGE     ?= ghcr.io/c3xdev/c3x:latest
TF_PLATFORM   ?= linux/amd64

# Selection
TF_OPTION     ?= option-b-lambda-functionurl
TF_ENV        ?= stage
AWS_REGION    ?= us-east-1
AWS_PROFILE   ?=

# Paths (Makefile is inside terraform/)
TF_WORKSPACE  := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
TF_ROOT       := $(TF_OPTION)
TF_ENV_DIR    := $(TF_ROOT)/env/$(TF_ENV)
TF_VARS       := env/$(TF_ENV)/terraform.tfvars
TF_BACKEND    := env/$(TF_ENV)/backend.hcl
TF_USAGE      := $(TF_ROOT)/env/$(TF_ENV)/c3x-usage.yml
TF_PLAN       := tfplan

# Optional: set TF_REFRESH=false to pass -refresh=false to plan.
TF_REFRESH    ?= true

# Docker arguments
DOCKER_COMMON = --rm --platform $(TF_PLATFORM) \
	-v "$(TF_WORKSPACE):/workspace" \
	-e AWS_ACCESS_KEY_ID \
	-e AWS_SECRET_ACCESS_KEY \
	-e AWS_SESSION_TOKEN \
	-e AWS_DEFAULT_REGION=$(AWS_REGION) \
	-e AWS_PROFILE \
	-e AWS_EC2_METADATA_DISABLED=true

# Mount AWS config for profiles / SSO.
ifneq ($(wildcard $(HOME)/.aws),)
DOCKER_COMMON += -v "$(HOME)/.aws:/root/.aws:ro"
endif

# OpenTofu image has tofu as its entrypoint.
TF_RUN = docker run $(DOCKER_COMMON) \
	-w "/workspace/$(TF_ROOT)" \
	$(TF_IMAGE)

# Interactive runner for apply/destroy.
TF_RUN_INTERACTIVE = docker run -it $(DOCKER_COMMON) \
	-w "/workspace/$(TF_ROOT)" \
	$(TF_IMAGE)

# C3X runs from the terraform directory.
C3X_RUN = docker run $(DOCKER_COMMON) \
	-w /workspace \
	$(C3X_IMAGE)

.PHONY: tf-init tf-validate tf-plan tf-show tf-show-json \
	tf-apply tf-destroy tf-output tf-fmt tf-estimate \
	tf-clean tf-show-config

tf-init:
	$(TF_RUN) init \
		-reconfigure \
		-backend-config=$(TF_BACKEND)

tf-validate:
	$(TF_RUN) validate

tf-plan:
	$(TF_RUN) plan \
		-var-file=$(TF_VARS) \
		$(if $(filter false,$(TF_REFRESH)),-refresh=false,) \
		-out=$(TF_PLAN)

tf-show:
	$(TF_RUN) show $(TF_PLAN)

tf-show-json:
	$(TF_RUN) show -json $(TF_PLAN) > "$(TF_WORKSPACE)/$(TF_ROOT)/plan.json"

tf-apply:
	$(TF_RUN_INTERACTIVE) apply $(TF_PLAN)

tf-destroy:
	$(TF_RUN_INTERACTIVE) destroy -var-file=$(TF_VARS)

tf-output:
	$(TF_RUN) output

tf-fmt:
	docker run --rm --platform $(TF_PLATFORM) \
		-v "$(TF_WORKSPACE):/workspace" \
		-w /workspace \
		$(TF_IMAGE) fmt -recursive .

## C3X
tf-estimate:
	c3x estimate \
		--path $(TF_ROOT) \
		--usage $(TF_USAGE) \
		--format csv > "$(TF_WORKSPACE)/$(TF_ENV_DIR)/cost.csv"

## Cleanup
tf-clean:
	rm -f "$(TF_WORKSPACE)/$(TF_ROOT)/tfplan" \
		"$(TF_WORKSPACE)/$(TF_ROOT)/plan.json" \
		"$(TF_WORKSPACE)/$(TF_ROOT)/cost.csv"

## Debug
tf-show-config:
	@echo "TF_OPTION    = $(TF_OPTION)"
	@echo "TF_ENV       = $(TF_ENV)"
	@echo "TF_ROOT      = $(TF_ROOT)"
	@echo "TF_VARS      = $(TF_VARS)"
	@echo "TF_BACKEND   = $(TF_BACKEND)"
	@echo "TF_USAGE     = $(TF_USAGE)"
	@echo "AWS_REGION   = $(AWS_REGION)"
	@echo "AWS_PROFILE  = $(AWS_PROFILE)"
	@echo "TF_IMAGE     = $(TF_IMAGE)"
	@echo "C3X_IMAGE    = $(C3X_IMAGE)"
	@echo "TF_PLATFORM  = $(TF_PLATFORM)"
	@echo "TF_REFRESH   = $(TF_REFRESH)"
