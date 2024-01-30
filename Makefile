KO_DOCKER_REPO=quay.io/sbaumer/debug-monitor
build:
	ko build --local --bare

KO_DOCKER_REPO=quay.io/sbaumer/debug-monitor
deploy:
	ko resolve --bare -f deployment.yaml > release.yaml
	kubectl apply -f release.yaml
