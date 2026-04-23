include ../../../../Makefile.variables
include ../../../../common.env

TASK_FILE ?= dummy-task.md
BRANCH := dev
CLAUDE_CONFIG_DIR ?= $(HOME)/.claude-agent
AGENT_DIR ?= $(shell cd ../.. && pwd)/agent
MODEL ?= sonnet
ALLOWED_TOOLS ?= Read,Grep,Glob,Bash(git:*),Bash(gh:*),WebFetch

generate-dummy-task:
	@echo '---' > $(TASK_FILE)
	@echo 'status: in_progress' >> $(TASK_FILE)
	@echo 'phase: planning' >> $(TASK_FILE)
	@echo 'title: Dummy PR Review' >> $(TASK_FILE)
	@echo 'assignee: pr-reviewer-agent' >> $(TASK_FILE)
	@echo '---' >> $(TASK_FILE)
	@echo '' >> $(TASK_FILE)
	@echo 'Review the pull request at https://github.com/bborbe/errors/pull/1 and return a verdict.' >> $(TASK_FILE)
	@echo "generated $(TASK_FILE)"

run-dummy-task: $(TASK_FILE)
	@go run -mod=mod . \
		-task-file=$(TASK_FILE) \
		-branch=$(BRANCH) \
		-agent-dir=$(AGENT_DIR) \
		-claude-config-dir=$(CLAUDE_CONFIG_DIR) \
		-model=$(MODEL) \
		-allowed-tools="$(ALLOWED_TOOLS)" \
		-v=2
