up:
	@docker compose up

build:
	@docker compose up --build

down:
	@docker compose down
	@docker volume prune -a -f